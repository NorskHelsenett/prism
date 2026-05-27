<script>
	import { run } from 'svelte/legacy';
	import Avatar from '$lib/components/Avatar.svelte';
	import DeleteModal from '$lib/components/DeleteModal.svelte';
	import { goto } from '$app/navigation';
	import { Fetch } from '$lib/fetchUtil';

	/**
	 * @typedef {Object} Props
	 * @property {boolean} [reload]
	 */

	/** @type {Props} */
	let { reload = true } = $props();

	const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
	const DEFAULT_COLOR = '#206bc4';
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

	const today = new Date();
	const currentYear = today.getFullYear();
	const formattedToday = MONTHS[today.getMonth()];

	function daysInMonth(month, year) {
		return new Date(year, month, 0).getDate();
	}
	const totalDays = daysInMonth(today.getMonth() + 1, currentYear);
	const dayPercentage = (today.getDate() / totalDays) * 100;

	function toDateOnly(value) {
		if (!value || typeof value !== 'string') return '';
		const prefix = value.slice(0, 10);
		if (prefix === '0001-01-01') return '';
		return prefix;
	}

	async function fetchProjects() {
		const response = await Fetch('/api/project/all');
		projects = (response || [])
			.map((p) => ({
				id: p.ID,
				name: p.ProjectName,
				dateFrom: toDateOnly(p.StartDate),
				dateTo: toDateOnly(p.EndDate),
				color: p.Color || DEFAULT_COLOR,
				isBugBounty: p.IsBugBounty,
				clientEmail: p.ClientEmail || '',
				hackerName: p.HackerName || '',
				groupId: p.GroupID || null,
				sortOrder: p.SortOrder || 0
			}))
			.filter((p) => p.dateFrom && p.dateTo)
			.sort((a, b) => {
				if (a.sortOrder !== b.sortOrder) return a.sortOrder - b.sortOrder;
				return a.dateFrom < b.dateFrom ? -1 : a.dateFrom > b.dateFrom ? 1 : 0;
			});
	}

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
			if (prefs && Array.isArray(prefs.collapsedGroups)) {
				collapsedGroups = new Set(prefs.collapsedGroups);
			}
		} catch (err) {
			console.error('Error fetching user preferences:', err);
		}
	}

	async function savePreferences() {
		try {
			await Fetch('/api/profile/preferences', {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ collapsedGroups: Array.from(collapsedGroups) })
			});
		} catch (err) {
			console.error('Error saving preferences:', err);
		}
	}

	async function loadAll() {
		await Promise.all([fetchProjects(), fetchGroups(), fetchPreferences()]);
	}

	run(() => {
		if (!reload) {
			loadAll();
		}
	});

	let groupedView = $derived.by(() => {
		const ungrouped = [];
		const byId = new Map();
		for (const g of groups) byId.set(g.id, { group: g, items: [] });
		for (const p of projects) {
			if (p.groupId && byId.has(p.groupId)) {
				byId.get(p.groupId).items.push(p);
			} else {
				ungrouped.push(p);
			}
		}
		return {
			ungrouped,
			groups: Array.from(byId.values()).sort((a, b) => a.group.sortOrder - b.group.sortOrder)
		};
	});

	function toggleCollapse(groupId) {
		const next = new Set(collapsedGroups);
		if (next.has(groupId)) next.delete(groupId);
		else next.add(groupId);
		collapsedGroups = next;
		savePreferences();
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
			await loadAll();
		} catch (err) {
			console.error('Error deleting group:', err);
		}
	}

	async function assignProjectToGroup(projectId, groupId) {
		try {
			await Fetch(`/api/project/${projectId}/group`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ groupId })
			});
			const p = projects.find((x) => x.id === projectId);
			if (p) {
				p.groupId = groupId;
				projects = [...projects];
			}
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
				assignProjectToGroup(projectA.id, newGroup.id),
				assignProjectToGroup(projectB.id, newGroup.id)
			]);
			startEditGroup(newGroup);
		} catch (err) {
			console.error('Error creating group:', err);
		}
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

		if (target.groupId) {
			// Drop on a project already in a group → add dragged to same group
			await assignProjectToGroup(dragged.id, target.groupId);
		} else if (dragged.groupId) {
			// Dragged from a group onto an ungrouped project → add target to dragged's group
			await assignProjectToGroup(target.id, dragged.groupId);
		} else {
			// Both ungrouped → create new group with both
			await createGroupFromProjects(dragged, target);
		}
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
		// groupId can be null (Ungrouped target)
		if (dragged.groupId === groupId) return;
		await assignProjectToGroup(dragged.id, groupId);
	}

	function eventIn(month, dateFrom, dateTo) {
		const monthIndex = MONTHS.indexOf(month);
		const fromDate = new Date(dateFrom);
		const toDate = new Date(dateTo);
		return fromDate.getMonth() <= monthIndex && toDate.getMonth() >= monthIndex;
	}

	function calculateStartPercentage(dateFrom, monthStr) {
		const date = new Date(dateFrom);
		const dateMonth = date.getMonth();
		const givenMonth = new Date(Date.parse(monthStr + ' 1, ' + date.getFullYear())).getMonth();
		if (dateMonth !== givenMonth) return 0;
		const total = new Date(date.getFullYear(), dateMonth + 1, 0).getDate();
		return (date.getDate() / total) * 100;
	}

	function calculateLineLength(dateTo, monthStr) {
		const date = new Date(dateTo);
		const dateMonth = date.getMonth();
		const givenMonth = new Date(Date.parse(monthStr + ' 1, ' + date.getFullYear())).getMonth();
		if (dateMonth !== givenMonth) return 100;
		const total = new Date(date.getFullYear(), dateMonth + 1, 0).getDate();
		return ((date.getDate() / total) * 100) * 0.8;
	}

	function splitEmails(value) {
		if (!value) return [];
		return value.split(',').map((s) => s.trim()).filter(Boolean);
	}

	function formatDate(d) {
		if (!d) return '—';
		const dt = new Date(d);
		return dt.toLocaleDateString(undefined, { day: '2-digit', month: 'short', year: 'numeric' });
	}
</script>

<div class="card">
	<div class="table-responsive small">
		<table class="table table-vcenter card-table">
			<thead>
				<tr>
					<th class="sticky-col first-col">Project</th>
					<th>Client</th>
					<th>Hackers</th>
					<th>Start</th>
					<th>End</th>
					<th>Type</th>
					{#each MONTHS as month}
						<th>{month}</th>
					{/each}
				</tr>
			</thead>
			<tbody>
				<!-- Ungrouped section -->
				<tr
					class="group-header ungrouped-header"
					class:drag-over={dragOverGroupKey === UNGROUPED_KEY}
					ondragover={(e) => onGroupHeaderDragOver(e, UNGROUPED_KEY)}
					ondragleave={() => onGroupHeaderDragLeave(UNGROUPED_KEY)}
					ondrop={(e) => onGroupHeaderDrop(e, null)}
				>
					<td colspan="18">
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

				<!-- Group sections -->
				{#each groupedView.groups as { group, items } (group.id)}
					{@const collapsed = collapsedGroups.has(group.id)}
					<tr
						class="group-header"
						class:drag-over={dragOverGroupKey === group.id}
						ondragover={(e) => onGroupHeaderDragOver(e, group.id)}
						ondragleave={() => onGroupHeaderDragLeave(group.id)}
						ondrop={(e) => onGroupHeaderDrop(e, group.id)}
					>
						<td colspan="18">
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

				{#if projects.length === 0}
					<tr>
						<td colspan="18" class="text-center text-secondary py-4">
							No projects with start/end dates yet. Set the dates on a project to see it on the timeline.
						</td>
					</tr>
				{/if}
			</tbody>
		</table>
	</div>
	<div class="card-footer d-flex align-items-center">
		<p class="m-0 text-secondary">
			Showing <span>{projects?.length || 0}</span> projects in <span>{groups.length}</span> groups · drag a row onto another to group them
		</p>
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
		<td class="sticky-col first-col" style="min-width:20em">
			<div class="project-name-cell">
				<span class="drag-grip" title="Drag to group">⋮⋮</span>
				<span class="project-color-dot" style="background-color: {project.color}"></span>
				<h4 class="text-capitalize m-0">
					<button class="link" onclick={() => goto(`/project/${project.id}/view`)}>
						{project.name}
					</button>
				</h4>
			</div>
		</td>
		<td>
			<div class="avatar-list avatar-list-stacked" style="min-width:6em">
				{#each splitEmails(project.clientEmail).slice(0, 4) as email}
					<Avatar email={email} option={{ showName: false, size: 'sm', emptyFields: true, circle: true, tooltipEnabled: false }} />
				{/each}
			</div>
		</td>
		<td>
			<div class="avatar-list avatar-list-stacked" style="min-width:6em">
				{#each splitEmails(project.hackerName).slice(0, 4) as email}
					<Avatar email={email} option={{ showName: false, size: 'sm', emptyFields: true, circle: true, tooltipEnabled: false }} />
				{/each}
			</div>
		</td>
		<td class="text-secondary text-nowrap">{formatDate(project.dateFrom)}</td>
		<td class="text-secondary text-nowrap">{formatDate(project.dateTo)}</td>
		<td>
			{#if project.isBugBounty}
				<span class="badge bg-yellow-lt">Bug Bounty</span>
			{:else}
				<span class="badge bg-cyan-lt">Project</span>
			{/if}
		</td>
		{#each MONTHS as month}
			{#if eventIn(month, project.dateFrom, project.dateTo)}
				<td class="timeline-container" class:today={month === formattedToday}>
					<span
						class="line"
						style="--start-percentage: {calculateStartPercentage(project.dateFrom, month)}%; --line-length: {calculateLineLength(project.dateTo, month)}%; --bar-color: {project.color};"
					></span>
					<span class:today-line={month === formattedToday} style="--day-percentage: {dayPercentage}"></span>
				</td>
			{:else}
				<td class:today={month === formattedToday}>
					<span class:today-line={month === formattedToday} style="--day-percentage: {dayPercentage}"></span>
				</td>
			{/if}
		{/each}
	</tr>
{/snippet}

<style>
	.table {
		overflow-x: hidden;
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

	.today-line {
		position: absolute;
		left: calc(var(--day-percentage) * 1%);
		width: 2px;
		background-color: rgba(var(--tblr-azure-rgb), 0.3);
		z-index: 1000;
		top: 0;
		height: 101%;
		bottom: 0;
	}

	.sticky-col h4 {
		margin-bottom: 0;
	}

	.sticky-col {
		position: -webkit-sticky;
		position: sticky;
		background-color: var(--tblr-body-bg);
		left: 0;
		z-index: 100;
	}

	.first-col {
		border-right: solid 1px var(--tblr-body-bg);
	}

	.line {
		position: absolute;
		bottom: 5px;
		left: var(--start-percentage);
		width: calc(var(--line-length) * 1.15);
		height: 10px;
		background-color: var(--bar-color, var(--tblr-azure));
		border-radius: 5px;
		bottom: 40%;
	}

	.timeline-container {
		position: relative;
	}

	.today {
		position: relative;
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
