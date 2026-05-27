<script>
	import { goto } from '$app/navigation';
	import { Fetch } from '$lib/fetchUtil.js';
	import Avatar from '../Avatar.svelte';
	import DeleteModal from '../DeleteModal.svelte';

	const UNGROUPED_KEY = 'ungrouped';

	let projects = $state([]);
	let groups = $state([]);
	let collapsedGroups = $state(new Set());
	let selectedRow = $state(-1);
	let editingGroupId = $state(null);
	let editingGroupName = $state('');
	let dragOverProjectId = $state(null);
	let dragOverGroupKey = $state(null);
	let draggingProjectId = $state(null);
	let showDeleteModal = $state(false);
	let pendingDeleteGroup = $state(null);
	let assignedToMe = $state(true);
	let activeOnly = $state(true);
	let searchQuery = $state('');
	let currentUserEmail = $state('');

	const todayISO = new Date().toISOString().split('T')[0];

	function toDateOnly(value) {
		if (!value || typeof value !== 'string') return '';
		const prefix = value.slice(0, 10);
		if (prefix === '0001-01-01') return '';
		return prefix;
	}

	async function fetchCurrentUser() {
		try {
			const me = await Fetch('/api/profile');
			if (me && me.email) currentUserEmail = me.email;
		} catch (err) {
			console.error('Error fetching current user:', err);
		}
	}

	function formatDate(dateString) {
		const date = new Date(dateString);
		return date
			.toLocaleDateString('en-CA', {
				year: 'numeric',
				month: '2-digit',
				day: '2-digit'
			})
			.replace(/-/g, '.');
	}

	async function getTotalVulnerabilites(projectID) {
		const total = await Fetch(`/api/project/${projectID}/vulnerabilities/total`);
		if (total) return total.total_vulnerabilities;
		return 0;
	}

	async function fetchProjects() {
		const response = await Fetch('/api/project/all');
		projects = (response || [])
			.map((p) => ({
				id: p.ID,
				name: p.ProjectName,
				createdAt: p.CreatedAt,
				clientEmail: p.ClientEmail || '',
				hackerName: p.HackerName || '',
				startDate: toDateOnly(p.StartDate),
				endDate: toDateOnly(p.EndDate),
				isBugBounty: p.IsBugBounty,
				color: p.Color || '',
				groupId: p.GroupID || null,
				sortOrder: p.SortOrder || 0
			}))
			.sort((a, b) => {
				if (a.sortOrder !== b.sortOrder) return a.sortOrder - b.sortOrder;
				return a.id - b.id;
			});
	}

	function isMineFn(email, project) {
		if (!email) return true;
		const lower = email.toLowerCase();
		const haystack = (project.clientEmail + ',' + project.hackerName).toLowerCase();
		return haystack.split(',').map((s) => s.trim()).includes(lower);
	}

	function isActiveFn(project) {
		// Projects without an end date are treated as active (ongoing or unscheduled).
		if (!project.endDate) return true;
		return project.endDate >= todayISO;
	}

	function matchesQuery(query, project) {
		if (!query) return true;
		const q = query.toLowerCase();
		return (
			project.name.toLowerCase().includes(q) ||
			project.clientEmail.toLowerCase().includes(q) ||
			project.hackerName.toLowerCase().includes(q)
		);
	}

	let filteredProjects = $derived(
		projects.filter((p) => {
			if (assignedToMe && !isMineFn(currentUserEmail, p)) return false;
			if (activeOnly && !isActiveFn(p)) return false;
			if (!matchesQuery(searchQuery, p)) return false;
			return true;
		})
	);

	async function fetchGroups() {
		const response = await Fetch('/api/project-group');
		groups = (response || []).map((g) => ({
			id: g.ID,
			name: g.Name,
			color: g.Color,
			sortOrder: g.SortOrder
		}));
	}

	async function fetchPreferences() {
		try {
			const prefs = await Fetch('/api/profile/preferences');
			if (!prefs) return;
			if (Array.isArray(prefs.collapsedGroups)) {
				collapsedGroups = new Set(prefs.collapsedGroups);
			}
			// Inverted persistence: defaults are "Assigned to me" and "Active only" both ON,
			// so we persist the *opt-out* flags. Absent / false flag → filter stays on.
			if (prefs.projectShowAll === true) assignedToMe = false;
			if (prefs.projectShowInactive === true) activeOnly = false;
		} catch (err) {
			console.error('Error fetching user preferences:', err);
		}
	}

	async function saveCollapsedGroups() {
		try {
			await Fetch('/api/profile/preferences', {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ collapsedGroups: Array.from(collapsedGroups) })
			});
		} catch (err) {
			console.error('Error saving collapsed groups:', err);
		}
	}

	async function saveFilters() {
		try {
			await Fetch('/api/profile/preferences', {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					projectShowAll: !assignedToMe,
					projectShowInactive: !activeOnly
				})
			});
		} catch (err) {
			console.error('Error saving filters:', err);
		}
	}

	function onAssignedToMeChange() {
		saveFilters();
	}

	function onActiveOnlyChange() {
		saveFilters();
	}

	export async function refreshList() {
		await Promise.all([
			fetchProjects(),
			fetchGroups(),
			fetchPreferences(),
			fetchCurrentUser()
		]);
	}

	let groupedView = $derived.by(() => {
		const ungrouped = [];
		const byId = new Map();
		for (const g of groups) byId.set(g.id, { group: g, items: [] });
		for (const p of filteredProjects) {
			if (p.groupId && byId.has(p.groupId)) {
				byId.get(p.groupId).items.push(p);
			} else {
				ungrouped.push(p);
			}
		}
		return {
			ungrouped,
			groups: Array.from(byId.values()).sort(
				(a, b) => a.group.sortOrder - b.group.sortOrder
			)
		};
	});

	function toggleCollapse(groupId) {
		const next = new Set(collapsedGroups);
		if (next.has(groupId)) next.delete(groupId);
		else next.add(groupId);
		collapsedGroups = next;
		saveCollapsedGroups();
	}

	function startEditGroup(group) {
		editingGroupId = group.id;
		editingGroupName = group.name;
	}

	async function commitGroupEdit() {
		if (editingGroupId == null) return;
		const id = editingGroupId;
		const newName = editingGroupName.trim() || 'Untitled group';
		const target = groups.find((g) => g.id === id);
		editingGroupId = null;
		if (!target || target.name === newName) return;
		target.name = newName;
		groups = [...groups];
		try {
			await Fetch(`/api/project-group/${id}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					name: newName,
					color: target.color || '',
					sortOrder: target.sortOrder || 0
				})
			});
		} catch (err) {
			console.error('Error renaming group:', err);
		}
	}

	function cancelGroupEdit() {
		editingGroupId = null;
	}

	async function createEmptyGroup() {
		try {
			const created = await Fetch('/api/project-group', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ name: 'New group', color: '', sortOrder: groups.length })
			});
			if (!created || !created.ID) return;
			const newGroup = {
				id: created.ID,
				name: created.Name,
				color: created.Color,
				sortOrder: created.SortOrder
			};
			groups = [...groups, newGroup];
			startEditGroup(newGroup);
		} catch (err) {
			console.error('Error creating group:', err);
		}
	}

	function requestDeleteGroup(group) {
		pendingDeleteGroup = group;
		showDeleteModal = true;
	}

	async function confirmDeleteGroup() {
		const group = pendingDeleteGroup;
		pendingDeleteGroup = null;
		if (!group) return;
		try {
			await Fetch(`/api/project-group/${group.id}`, { method: 'DELETE' });
			await refreshList();
		} catch (err) {
			console.error('Error deleting group:', err);
		}
	}

	async function assignProjectToGroup(projectId, groupId, sortOrder) {
		try {
			const body = { groupId };
			if (typeof sortOrder === 'number') body.sortOrder = sortOrder;
			await Fetch(`/api/project/${projectId}/group`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(body)
			});
		} catch (err) {
			console.error('Error assigning project to group:', err);
		}
	}

	async function createGroupFromProjects(projectA, projectB) {
		try {
			const created = await Fetch('/api/project-group', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ name: 'New group', color: '', sortOrder: groups.length })
			});
			if (!created || !created.ID) return;
			const newGroup = {
				id: created.ID,
				name: created.Name,
				color: created.Color,
				sortOrder: created.SortOrder
			};
			groups = [...groups, newGroup];
			await Promise.all([
				assignProjectToGroup(projectA.id, newGroup.id, 0),
				assignProjectToGroup(projectB.id, newGroup.id, 1)
			]);
			projectA.groupId = newGroup.id;
			projectA.sortOrder = 0;
			projectB.groupId = newGroup.id;
			projectB.sortOrder = 1;
			projects = [...projects];
			startEditGroup(newGroup);
		} catch (err) {
			console.error('Error creating group:', err);
		}
	}

	// Reorder dragged before target within the same section (or move dragged into target's section
	// at target's position). Returns the new array of {id, sortOrder} pairs that were updated, and
	// the new groupId for the dragged project.
	function computeReorder(dragged, target) {
		const targetGroupId = target.groupId || null;
		// Build current order of the destination section
		const sectionItems = projects
			.filter((p) => (p.groupId || null) === targetGroupId && p.id !== dragged.id)
			.sort((a, b) => a.sortOrder - b.sortOrder || a.id - b.id);
		const targetIndex = sectionItems.findIndex((p) => p.id === target.id);
		const insertAt = targetIndex >= 0 ? targetIndex : sectionItems.length;
		sectionItems.splice(insertAt, 0, { ...dragged, groupId: targetGroupId });
		// Renumber sortOrder for the section
		return sectionItems.map((p, i) => ({ id: p.id, sortOrder: i, groupId: targetGroupId }));
	}

	async function applyReorder(updates) {
		// Optimistically update local state
		const byId = new Map(updates.map((u) => [u.id, u]));
		projects = projects.map((p) => {
			const u = byId.get(p.id);
			if (!u) return p;
			return { ...p, sortOrder: u.sortOrder, groupId: u.groupId };
		});
		// Persist in parallel
		await Promise.all(
			updates.map((u) =>
				Fetch(`/api/project/${u.id}/group`, {
					method: 'PATCH',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({ groupId: u.groupId, sortOrder: u.sortOrder })
				})
			)
		);
	}

	function onProjectDragStart(e, project) {
		draggingProjectId = project.id;
		e.dataTransfer.effectAllowed = 'move';
		try {
			e.dataTransfer.setData('text/plain', String(project.id));
		} catch (_) {}
	}

	function onProjectDragEnd() {
		draggingProjectId = null;
		dragOverProjectId = null;
		dragOverGroupKey = null;
	}

	function onProjectDragOver(e, project) {
		if (draggingProjectId == null || draggingProjectId === project.id) return;
		e.preventDefault();
		e.dataTransfer.dropEffect = 'move';
		dragOverProjectId = project.id;
	}

	function onProjectDragLeave(project) {
		if (dragOverProjectId === project.id) dragOverProjectId = null;
	}

	async function onProjectDrop(e, target) {
		e.preventDefault();
		const draggedId = draggingProjectId ?? Number(e.dataTransfer.getData('text/plain'));
		dragOverProjectId = null;
		if (!draggedId || draggedId === target.id) return;
		const dragged = projects.find((p) => p.id === draggedId);
		if (!dragged) return;

		const draggedGroup = dragged.groupId || null;
		const targetGroup = target.groupId || null;

		if (draggedGroup === null && targetGroup === null) {
			// Both ungrouped → create new group with both
			await createGroupFromProjects(dragged, target);
			return;
		}
		// Reorder within same section, or move dragged into target's section at target's position
		const updates = computeReorder(dragged, target);
		await applyReorder(updates);
	}

	function onGroupHeaderDragOver(e, groupKey) {
		if (draggingProjectId == null) return;
		e.preventDefault();
		e.dataTransfer.dropEffect = 'move';
		dragOverGroupKey = groupKey;
	}

	function onGroupHeaderDragLeave(groupKey) {
		if (dragOverGroupKey === groupKey) dragOverGroupKey = null;
	}

	async function onGroupHeaderDrop(e, groupId) {
		e.preventDefault();
		const draggedId = draggingProjectId ?? Number(e.dataTransfer.getData('text/plain'));
		dragOverGroupKey = null;
		if (!draggedId) return;
		const dragged = projects.find((p) => p.id === draggedId);
		if (!dragged) return;
		if ((dragged.groupId || null) === groupId) return;
		// Place at end of section
		const sectionItems = projects
			.filter((p) => (p.groupId || null) === groupId && p.id !== dragged.id)
			.sort((a, b) => a.sortOrder - b.sortOrder || a.id - b.id);
		const newSortOrder = sectionItems.length;
		await assignProjectToGroup(dragged.id, groupId, newSortOrder);
		const updated = projects.find((p) => p.id === dragged.id);
		if (updated) {
			updated.groupId = groupId;
			updated.sortOrder = newSortOrder;
			projects = [...projects];
		}
	}

	function splitEmails(value) {
		if (!value) return [];
		return value.split(',').map((s) => s.trim()).filter(Boolean);
	}
</script>

<div class="card mt-3">
	<div class="card-header">
		<div class="card-header-content">
			<div class="search-row">
				<div class="input-icon flex-grow-1">
					<span class="input-icon-addon">
						<svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
							<path stroke="none" d="M0 0h24v24H0z" fill="none"/>
							<path d="M10 10m-7 0a7 7 0 1 0 14 0a7 7 0 1 0 -14 0"/>
							<path d="M21 21l-6 -6"/>
						</svg>
					</span>
					<input
						type="text"
						class="form-control"
						placeholder="Search by project name, client, or hacker..."
						bind:value={searchQuery}
					/>
				</div>
				<label class="form-check form-switch mb-0 ms-2 toggle-nowrap">
					<input
						class="form-check-input"
						type="checkbox"
						bind:checked={assignedToMe}
						onchange={onAssignedToMeChange}
					/>
					<span class="form-check-label">Assigned to me</span>
				</label>
				<label class="form-check form-switch mb-0 ms-2 toggle-nowrap">
					<input
						class="form-check-input"
						type="checkbox"
						bind:checked={activeOnly}
						onchange={onActiveOnlyChange}
					/>
					<span class="form-check-label">Active only</span>
				</label>
				<button class="btn btn-sm btn-outline-primary ms-2 toggle-nowrap" type="button" onclick={createEmptyGroup}>
					<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M12 5l0 14"/><path d="M5 12l14 0"/></svg>
					New group
				</button>
			</div>
		</div>
	</div>
	<div class="table-responsive col-12">
		<table class="table table-vcenter card-table">
			<tbody>
				<tr
					class="group-header ungrouped-header"
					class:drag-over={dragOverGroupKey === UNGROUPED_KEY}
					ondragover={(e) => onGroupHeaderDragOver(e, UNGROUPED_KEY)}
					ondragleave={() => onGroupHeaderDragLeave(UNGROUPED_KEY)}
					ondrop={(e) => onGroupHeaderDrop(e, null)}
				>
					<td colspan="4">
						<div class="group-header-inner">
							<span class="group-caret">·</span>
							<span class="group-name muted">Ungrouped</span>
							<span class="group-count">{groupedView.ungrouped.length}</span>
						</div>
					</td>
				</tr>
				{#each groupedView.ungrouped as project (project.id)}
					{@render projectRow(project)}
				{/each}

				{#each groupedView.groups as { group, items } (group.id)}
					{@const collapsed = collapsedGroups.has(group.id)}
					<tr
						class="group-header"
						class:drag-over={dragOverGroupKey === group.id}
						ondragover={(e) => onGroupHeaderDragOver(e, group.id)}
						ondragleave={() => onGroupHeaderDragLeave(group.id)}
						ondrop={(e) => onGroupHeaderDrop(e, group.id)}
					>
						<td colspan="4">
							<div class="group-header-inner">
								<button
									type="button"
									class="group-caret-btn"
									onclick={() => toggleCollapse(group.id)}
									title={collapsed ? 'Expand' : 'Collapse'}
								>
									<span class="group-caret" class:collapsed>▾</span>
								</button>
								{#if editingGroupId === group.id}
									<input
										class="group-name-input"
										bind:value={editingGroupName}
										onblur={commitGroupEdit}
										onkeydown={(e) => {
											if (e.key === 'Enter') commitGroupEdit();
											else if (e.key === 'Escape') cancelGroupEdit();
										}}
										autofocus
									/>
								{:else}
									<button
										type="button"
										class="group-name-btn"
										ondblclick={() => startEditGroup(group)}
										title="Double-click to rename"
									>
										{group.name}
									</button>
								{/if}
								<span class="group-count">{items.length}</span>
								<button
									type="button"
									class="group-delete-btn"
									onclick={() => requestDeleteGroup(group)}
									title="Delete group"
								>
									<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M4 7l16 0"/><path d="M10 11l0 6"/><path d="M14 11l0 6"/><path d="M5 7l1 12a2 2 0 0 0 2 2h8a2 2 0 0 0 2 -2l1 -12"/><path d="M9 7v-3a1 1 0 0 1 1 -1h4a1 1 0 0 1 1 1v3"/></svg>
								</button>
							</div>
						</td>
					</tr>
					{#if !collapsed}
						{#each items as project (project.id)}
							{@render projectRow(project)}
						{/each}
					{/if}
				{/each}

				{#if filteredProjects.length === 0}
					<tr>
						<td colspan="4" class="text-center text-secondary py-4">
							{#if projects.length === 0}
								No projects yet.
							{:else}
								No projects match the current filters.
							{/if}
						</td>
					</tr>
				{/if}
			</tbody>
		</table>
	</div>
</div>

<DeleteModal
	bind:showDeleteModal
	onDelete={confirmDeleteGroup}
	deleteButtonText="Delete group"
	accent="danger"
>
	{#if pendingDeleteGroup}
		<div class="text-secondary">
			Delete <strong>{pendingDeleteGroup.name}</strong>? Its projects will move back to Ungrouped — they won't be deleted.
		</div>
	{/if}
</DeleteModal>

{#snippet projectRow(project)}
	{@const clients = splitEmails(project.clientEmail)}
	{@const hackers = splitEmails(project.hackerName)}
	<tr
		draggable="true"
		ondragstart={(e) => onProjectDragStart(e, project)}
		ondragend={onProjectDragEnd}
		ondragover={(e) => onProjectDragOver(e, project)}
		ondragleave={() => onProjectDragLeave(project)}
		ondrop={(e) => onProjectDrop(e, project)}
		ondblclick={() => goto(`/project/${project.id}/view`)}
		onclick={() => (selectedRow === project.id ? (selectedRow = -1) : (selectedRow = project.id))}
		class:selected={selectedRow === project.id}
		class:drag-over={dragOverProjectId === project.id}
		class:dragging={draggingProjectId === project.id}
	>
		<td class="project-cell">
			<div class="project-card">
				<div class="project-card-top">
					<span class="drag-grip" title="Drag to group or reorder">⋮⋮</span>
					{#if project.color}
						<span class="project-color-dot" style="background-color: {project.color}"></span>
					{/if}
					<button class="link project-name" onclick={() => goto(`/project/${project.id}/view`)}>
						{project.name}
					</button>
					{#if project.isBugBounty}
						<span class="badge badge-outline text-azure ms-1">
							<svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-bug" width="16" height="16" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M9 9v-1a3 3 0 0 1 6 0v1"/><path d="M8 9h8a6 6 0 0 1 1 3v3a5 5 0 0 1 -10 0v-3a6 6 0 0 1 1 -3"/><path d="M3 13l4 0"/><path d="M17 13l4 0"/><path d="M12 20l0 -6"/><path d="M4 19l3.35 -2"/><path d="M20 19l-3.35 -2"/><path d="M4 7l3.75 2.4"/><path d="M20 7l-3.75 2.4"/></svg>
							Bug Bounty
						</span>
					{/if}
				</div>
				<div class="project-card-meta">
					<div class="meta-item" title="Clients">
						<svg xmlns="http://www.w3.org/2000/svg" class="icon icon-sm text-secondary" width="16" height="16" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M3 7m0 2a2 2 0 0 1 2 -2h14a2 2 0 0 1 2 2v9a2 2 0 0 1 -2 2h-14a2 2 0 0 1 -2 -2z"/><path d="M8 7v-2a2 2 0 0 1 2 -2h4a2 2 0 0 1 2 2v2"/><path d="M12 12l0 .01"/><path d="M3 13a20 20 0 0 0 18 0"/></svg>
						<span class="meta-label">Clients</span>
						{#if clients.length}
							<div class="avatar-list avatar-list-stacked">
								{#each clients.slice(0, 6) as email (email)}
									<Avatar email={email} option={{ showName: false, emptyFields: true, size: 'xs', circle: true }} />
								{/each}
								{#if clients.length > 6}
									<span class="avatar avatar-xs rounded-circle">+{clients.length - 6}</span>
								{/if}
							</div>
						{:else}
							<span class="meta-empty">none</span>
						{/if}
					</div>
					<div class="meta-item" title="Assigned hackers">
						<svg xmlns="http://www.w3.org/2000/svg" class="icon icon-sm text-secondary" width="16" height="16" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M9 12a3 3 0 1 0 6 0a3 3 0 0 0 -6 0"/><path d="M17.657 16.657l-4.243 4.243a2 2 0 0 1 -2.827 0l-4.244 -4.243a8 8 0 1 1 11.314 0z"/></svg>
						<span class="meta-label">Hackers</span>
						{#if hackers.length}
							<div class="avatar-list avatar-list-stacked">
								{#each hackers.slice(0, 6) as email (email)}
									<Avatar email={email} option={{ showName: false, emptyFields: true, size: 'xs', circle: true }} />
								{/each}
								{#if hackers.length > 6}
									<span class="avatar avatar-xs rounded-circle">+{hackers.length - 6}</span>
								{/if}
							</div>
						{:else}
							<span class="meta-empty">none</span>
						{/if}
					</div>
				</div>
			</div>
		</td>
		<td class="dates-cell">
			<div class="date-stack">
				<svg class="date-icon icon icon-tabler icon-tabler-calendar text-secondary" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M4 7a2 2 0 0 1 2 -2h12a2 2 0 0 1 2 2v12a2 2 0 0 1 -2 2h-12a2 2 0 0 1 -2 -2v-12z"/><path d="M16 3v4"/><path d="M8 3v4"/><path d="M4 11h16"/><path d="M11 15h1"/><path d="M12 15v3"/></svg>
				<div class="date-rows">
					{#if project.startDate || project.endDate}
						<span class="date-label">Start</span>
						<span class="date-value">{project.startDate || '—'}</span>
						<span class="date-label">End</span>
						<span class="date-value">{project.endDate || '—'}</span>
					{:else}
						<span class="date-label">Created</span>
						<span class="date-value text-secondary">{formatDate(project.createdAt)}</span>
					{/if}
				</div>
			</div>
		</td>
		<td class="vulns-cell">
			<div class="text-secondary fw-bold text-end d-inline-flex align-items-center gap-1">
				<svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-shield-half-filled" width="20" height="20" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M12 3a12 12 0 0 0 8.5 3a12 12 0 0 1 -8.5 15a12 12 0 0 1 -8.5 -15a12 12 0 0 0 8.5 -3"/><path d="M12 3v18"/><path d="M12 11h8.9"/><path d="M12 8h8.9"/><path d="M12 5h3.1"/><path d="M12 17h6.2"/><path d="M12 14h8"/></svg>
				{#await getTotalVulnerabilites(project.id)}
				{:then totalVulnerabilities}
					{totalVulnerabilities}
				{/await}
			</div>
		</td>
		<td class="view-cell">
			<a href={`/project/${project.id}/view`}>View</a>
		</td>
	</tr>
{/snippet}

<style>
	.card-header {
		border-bottom: none;
		padding-bottom: 0.2rem;
	}

	.card-header-content {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		width: 100%;
		padding: 0.25rem 0;
	}

	.search-row {
		display: flex;
		align-items: center;
		gap: 1rem;
		flex-wrap: wrap;
	}

	.toggle-nowrap {
		white-space: nowrap;
		cursor: pointer;
	}

	.card-header-content :global(.form-check-input) {
		cursor: pointer;
	}

	.selected {
		background-color: rgba(184, 196, 228, 0.05);
		cursor: pointer;
	}

	tr.dragging {
		opacity: 0.4;
	}

	tr.drag-over > td {
		background-color: rgba(32, 107, 196, 0.12);
		box-shadow: inset 0 -2px 0 var(--tblr-primary);
	}

	.project-cell {
		min-width: 24em;
		padding-top: 10px;
		padding-bottom: 10px;
	}

	.project-card {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.project-card-top {
		display: flex;
		align-items: center;
		gap: 8px;
		min-width: 0;
	}

	.project-name {
		font-weight: 600;
		font-size: 0.95rem;
		text-overflow: ellipsis;
		overflow: hidden;
		white-space: nowrap;
		min-width: 0;
	}

	.project-card-meta {
		display: flex;
		align-items: center;
		gap: 18px;
		flex-wrap: wrap;
		padding-left: 28px;
	}

	.meta-item {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-size: 0.8rem;
		color: var(--tblr-body-color);
		min-width: 0;
	}

	.meta-label {
		font-size: 0.7rem;
		font-weight: 500;
		color: var(--tblr-muted);
		text-transform: uppercase;
		letter-spacing: 0.04em;
	}

	.meta-empty {
		font-size: 0.75rem;
		color: var(--tblr-muted);
		font-style: italic;
	}

	.meta-item.dates .meta-value {
		font-size: 0.78rem;
		font-variant-numeric: tabular-nums;
	}

	.dates-cell {
		width: 1%;
		white-space: nowrap;
		padding-left: 16px;
		padding-right: 16px;
		vertical-align: middle;
	}

	.date-stack {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	.date-icon {
		flex-shrink: 0;
	}

	.date-rows {
		display: grid;
		grid-template-columns: 3rem 7rem;
		row-gap: 2px;
		column-gap: 8px;
		font-size: 0.78rem;
		font-variant-numeric: tabular-nums;
		line-height: 1.2;
	}

	.date-label {
		text-align: right;
		font-size: 0.68rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.04em;
		color: var(--tblr-muted);
	}

	.date-value {
		color: var(--tblr-body-color);
	}

	.vulns-cell {
		width: 1%;
		white-space: nowrap;
		text-align: right;
		padding-right: 16px;
	}

	.view-cell {
		width: 1%;
		white-space: nowrap;
	}

	.project-name-cell {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.drag-grip {
		cursor: grab;
		color: var(--tblr-muted);
		opacity: 0.4;
		font-size: 0.9rem;
		letter-spacing: -2px;
		user-select: none;
	}

	tr:hover .drag-grip {
		opacity: 0.9;
	}

	.drag-grip:active {
		cursor: grabbing;
	}

	.project-color-dot {
		width: 12px;
		height: 12px;
		border-radius: 50%;
		flex-shrink: 0;
		box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.08);
	}

	.group-header > td {
		background-color: var(--tblr-bg-surface-secondary, rgba(0, 0, 0, 0.03));
		border-top: 1px solid var(--tblr-border-color);
		padding: 6px 12px;
	}

	.group-header.ungrouped-header > td {
		background-color: transparent;
		border-top: none;
	}

	.group-header.drag-over > td {
		background-color: rgba(32, 107, 196, 0.18);
		box-shadow: inset 0 0 0 2px var(--tblr-primary);
	}

	.group-header-inner {
		display: flex;
		align-items: center;
		gap: 10px;
		font-size: 0.85rem;
		font-weight: 600;
		color: var(--tblr-body-color);
	}

	.group-caret-btn {
		background: none;
		border: none;
		padding: 2px 4px;
		cursor: pointer;
		color: var(--tblr-muted);
	}

	.group-caret {
		display: inline-block;
		transition: transform 0.15s ease;
		font-size: 0.85rem;
	}

	.group-caret.collapsed {
		transform: rotate(-90deg);
	}

	.group-name-btn {
		background: none;
		border: none;
		padding: 2px 6px;
		font: inherit;
		font-weight: 600;
		color: inherit;
		cursor: text;
		border-radius: 3px;
	}

	.group-name-btn:hover {
		background-color: rgba(0, 0, 0, 0.05);
	}

	.group-name-input {
		background: var(--tblr-body-bg);
		border: 1px solid var(--tblr-primary);
		border-radius: 3px;
		padding: 2px 6px;
		font: inherit;
		font-weight: 600;
		outline: none;
		min-width: 120px;
	}

	.group-count {
		font-size: 0.7rem;
		font-weight: 500;
		color: var(--tblr-muted);
		background-color: rgba(0, 0, 0, 0.06);
		padding: 1px 8px;
		border-radius: 10px;
	}

	.group-name.muted {
		color: var(--tblr-muted);
		font-weight: 500;
	}

	.group-delete-btn {
		background: none;
		border: none;
		padding: 2px 4px;
		cursor: pointer;
		color: var(--tblr-muted);
		opacity: 0;
		transition: opacity 0.15s ease;
		margin-left: auto;
	}

	.group-header:hover .group-delete-btn {
		opacity: 1;
	}

	.group-delete-btn:hover {
		color: var(--tblr-danger);
	}

	.link {
		background: none;
		border: none;
		padding: 0;
		margin: 0;
		color: inherit;
		font: inherit;
		cursor: pointer;
		text-align: left;
	}

	.link:hover {
		color: rgba(var(--tblr-link-color-rgb), var(--tblr-link-opacity, 1));
	}
</style>
