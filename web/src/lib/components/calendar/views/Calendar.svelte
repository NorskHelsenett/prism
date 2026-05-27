<script>
	import { run } from 'svelte/legacy';
	import { onMount, onDestroy, tick } from 'svelte';
	import { Fetch } from '$lib/fetchUtil';
	import { goto } from '$app/navigation';

	/**
	 * @typedef {Object} Props
	 * @property {boolean} [reload]
	 */

	/** @type {Props} */
	let { reload = false } = $props();

	const MONTH_NAMES = [
		'January', 'February', 'March', 'April', 'May', 'June',
		'July', 'August', 'September', 'October', 'November', 'December'
	];
	const WEEKDAY_LABELS = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
	const DEFAULT_COLOR = '#206bc4';
	const MS_PER_DAY = 1000 * 60 * 60 * 24;

	let scrollerRef = $state();
	let topSentinel = $state();
	let bottomSentinel = $state();

	let monthsRange = $state([]);
	let projects = $state([]);
	let groups = $state([]);
	let collapsedGroups = $state(new Set());

	let topObs;
	let bottomObs;
	let isLoadingTop = false;
	let isLoadingBottom = false;
	let didInitialScroll = false;

	const todayDate = new Date();
	todayDate.setHours(0, 0, 0, 0);

	function monthKey(year, month) {
		return `${year}-${String(month).padStart(2, '0')}`;
	}

	function relMonth(date, offset) {
		const d = new Date(date.getFullYear(), date.getMonth() + offset, 1);
		return { key: monthKey(d.getFullYear(), d.getMonth()), year: d.getFullYear(), month: d.getMonth() };
	}

	function isoDate(d) {
		const year = d.getFullYear();
		const month = String(d.getMonth() + 1).padStart(2, '0');
		const day = String(d.getDate()).padStart(2, '0');
		return `${year}-${month}-${day}`;
	}

	function startOfWeekMonday(date) {
		const d = new Date(date);
		const wd = (d.getDay() + 6) % 7;
		d.setDate(d.getDate() - wd);
		d.setHours(0, 0, 0, 0);
		return d;
	}

	function isSameDay(a, b) {
		return a.getFullYear() === b.getFullYear()
			&& a.getMonth() === b.getMonth()
			&& a.getDate() === b.getDate();
	}

	function getWeeksForMonth(year, month) {
		const firstDay = new Date(year, month, 1);
		const lastDay = new Date(year, month + 1, 0);
		const startMonday = startOfWeekMonday(firstDay);
		const lastDayOfWeek = (lastDay.getDay() + 6) % 7;
		const endSunday = new Date(lastDay);
		endSunday.setDate(lastDay.getDate() + (6 - lastDayOfWeek));

		const weeks = [];
		const cursor = new Date(startMonday);
		while (cursor <= endSunday) {
			const week = [];
			for (let i = 0; i < 7; i++) {
				week.push(new Date(cursor));
				cursor.setDate(cursor.getDate() + 1);
			}
			weeks.push(week);
		}
		return weeks;
	}

	function toDateOnly(value) {
		if (!value || typeof value !== 'string') return '';
		const prefix = value.slice(0, 10);
		if (prefix === '0001-01-01') return '';
		return prefix;
	}

	async function fetchProjects() {
		// Fetch all projects with dates — client-side filters by visible months.
		const result = await Fetch('/api/project/all');
		const list = (result || [])
			.map((p) => ({
				id: p.ID,
				name: p.ProjectName,
				dateFrom: toDateOnly(p.StartDate),
				dateTo: toDateOnly(p.EndDate),
				color: p.Color || DEFAULT_COLOR,
				isBugBounty: p.IsBugBounty,
				clientEmail: p.ClientEmail || '',
				groupId: p.GroupID || null
			}))
			.filter((p) => p.dateFrom && p.dateTo);
		projects = list;
	}

	async function fetchGroups() {
		const result = await Fetch('/api/project-group');
		groups = (result || []).map((g) => ({
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
			console.error('Error fetching preferences:', err);
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

	function toggleGroup(groupId) {
		const next = new Set(collapsedGroups);
		if (next.has(groupId)) next.delete(groupId);
		else next.add(groupId);
		collapsedGroups = next;
		savePreferences();
	}

	let visibleProjects = $derived(
		projects.filter((p) => !p.groupId || !collapsedGroups.has(p.groupId))
	);

	async function loadMore(direction) {
		if (direction === 'top' && !isLoadingTop) {
			isLoadingTop = true;
			const first = monthsRange[0];
			const prev = relMonth(new Date(first.year, first.month, 1), -1);
			const prevHeight = scrollerRef?.scrollHeight ?? 0;
			const prevScrollTop = scrollerRef?.scrollTop ?? 0;
			monthsRange = [prev, ...monthsRange];
			await tick();
			if (scrollerRef) {
				const newHeight = scrollerRef.scrollHeight;
				scrollerRef.scrollTop = prevScrollTop + (newHeight - prevHeight);
			}
			isLoadingTop = false;
		} else if (direction === 'bottom' && !isLoadingBottom) {
			isLoadingBottom = true;
			const last = monthsRange[monthsRange.length - 1];
			const next = relMonth(new Date(last.year, last.month, 1), 1);
			monthsRange = [...monthsRange, next];
			isLoadingBottom = false;
		}
	}

	onMount(async () => {
		const now = new Date();
		monthsRange = [relMonth(now, -1), relMonth(now, 0), relMonth(now, 1)];

		await Promise.all([fetchProjects(), fetchGroups(), fetchPreferences()]);
		await tick();

		if (scrollerRef && !didInitialScroll) {
			const cur = scrollerRef.querySelector('[data-current="true"]');
			if (cur) cur.scrollIntoView({ block: 'start' });
			didInitialScroll = true;
		}

		await tick();

		if (topSentinel && bottomSentinel && scrollerRef) {
			topObs = new IntersectionObserver(
				(entries) => {
					if (entries[0].isIntersecting && didInitialScroll) loadMore('top');
				},
				{ root: scrollerRef, rootMargin: '300px 0px 0px 0px' }
			);
			bottomObs = new IntersectionObserver(
				(entries) => {
					if (entries[0].isIntersecting && didInitialScroll) loadMore('bottom');
				},
				{ root: scrollerRef, rootMargin: '0px 0px 300px 0px' }
			);
			topObs.observe(topSentinel);
			bottomObs.observe(bottomSentinel);
		}
	});

	onDestroy(() => {
		topObs?.disconnect();
		bottomObs?.disconnect();
	});

	run(() => {
		if (!reload) {
			Promise.all([fetchProjects(), fetchGroups(), fetchPreferences()]);
		}
	});

	function computeWeeksWithBands(year, month, allProjects) {
		const weeks = getWeeksForMonth(year, month);
		return weeks.map((weekDays) => {
			const weekStart = weekDays[0];
			const weekEnd = new Date(weekDays[6]);
			weekEnd.setHours(23, 59, 59, 999);

			const intersecting = allProjects.filter((p) => {
				const s = new Date(p.dateFrom); s.setHours(0, 0, 0, 0);
				const en = new Date(p.dateTo); en.setHours(0, 0, 0, 0);
				return s <= weekEnd && en >= weekStart;
			});

			const bands = intersecting.map((p) => {
				const s = new Date(p.dateFrom); s.setHours(0, 0, 0, 0);
				const en = new Date(p.dateTo); en.setHours(0, 0, 0, 0);
				const startCol = Math.max(0, Math.round((s - weekStart) / MS_PER_DAY));
				const endCol = Math.min(6, Math.round((en - weekStart) / MS_PER_DAY));
				return {
					project: p,
					startCol,
					span: Math.max(1, endCol - startCol + 1),
					startsHere: s >= weekStart,
					endsHere: en <= weekEnd
				};
			});

			bands.sort((a, b) => {
				if (a.startCol !== b.startCol) return a.startCol - b.startCol;
				return b.span - a.span;
			});

			const lanes = [];
			for (const band of bands) {
				const bandEnd = band.startCol + band.span - 1;
				let placed = false;
				for (let i = 0; i < lanes.length; i++) {
					const conflict = lanes[i].some(
						(b) => band.startCol <= b.endCol && bandEnd >= b.startCol
					);
					if (!conflict) {
						band.lane = i;
						lanes[i].push({ startCol: band.startCol, endCol: bandEnd });
						placed = true;
						break;
					}
				}
				if (!placed) {
					band.lane = lanes.length;
					lanes.push([{ startCol: band.startCol, endCol: bandEnd }]);
				}
			}

			return { weekDays, bands, laneCount: Math.max(1, lanes.length) };
		});
	}

	let monthsView = $derived(
		monthsRange.map((m) => ({
			...m,
			weeks: computeWeeksWithBands(m.year, m.month, visibleProjects)
		}))
	);

	function isCurrentMonth(m) {
		const now = new Date();
		return m.year === now.getFullYear() && m.month === now.getMonth();
	}

	function openProject(project) {
		goto(`/project/${project.id}/view`);
	}

	function scrollToToday() {
		if (!scrollerRef) return;
		const cur = scrollerRef.querySelector('[data-current="true"]');
		if (cur) cur.scrollIntoView({ block: 'start', behavior: 'smooth' });
	}
</script>

<div class="calendar-shell card">
	<div class="calendar-toolbar">
		<button class="btn btn-sm btn-outline-primary" type="button" onclick={scrollToToday}>
			<svg xmlns="http://www.w3.org/2000/svg" class="icon" width="16" height="16" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M4 7a2 2 0 0 1 2 -2h12a2 2 0 0 1 2 2v12a2 2 0 0 1 -2 2h-12a2 2 0 0 1 -2 -2v-12z"/><path d="M16 3v4"/><path d="M8 3v4"/><path d="M4 11h16"/><path d="M11 15h1"/><path d="M12 15v3"/></svg>
			Today
		</button>
		{#if groups.length > 0}
			<div class="group-legend">
				{#each groups as group (group.id)}
					{@const collapsed = collapsedGroups.has(group.id)}
					<button
						type="button"
						class="group-chip"
						class:collapsed
						onclick={() => toggleGroup(group.id)}
						title={collapsed ? 'Show this group' : 'Hide this group'}
					>
						<span class="chip-dot" style="background-color: {group.color || '#94a3b8'}"></span>
						<span class="chip-name">{group.name}</span>
					</button>
				{/each}
			</div>
		{/if}
	</div>

	<div class="calendar-scroller" bind:this={scrollerRef}>
		<div bind:this={topSentinel} class="sentinel"></div>

		{#each monthsView as month (month.key)}
			<section class="month-section" data-current={isCurrentMonth(month)}>
				<header class="month-header">
					<h2 class="month-title">
						{MONTH_NAMES[month.month]}
						<span class="month-year">{month.year}</span>
					</h2>
				</header>

				<div class="weekday-header">
					{#each WEEKDAY_LABELS as label, i}
						<div class="weekday-cell" class:weekend={i >= 5}>{label}</div>
					{/each}
				</div>

				<div class="weeks">
					{#each month.weeks as week, wi (wi)}
						<div class="week-row" style="--lane-count: {week.laneCount}">
							<div class="day-grid">
								{#each week.weekDays as day, di}
									{@const inMonth = day.getMonth() === month.month}
									{@const isWeekend = di >= 5}
									{@const isToday = isSameDay(day, todayDate)}
									<div
										class="day-cell"
										class:out-of-month={!inMonth}
										class:weekend={isWeekend}
										class:today={isToday}
									>
										<div class="day-number">{day.getDate()}</div>
									</div>
								{/each}
							</div>

							<div class="project-band-layer">
								{#each week.bands as band (band.project.id + '-' + wi)}
									<!-- svelte-ignore a11y_click_events_have_key_events -->
									<button
										type="button"
										class="project-band projectid-{band.project.id}"
										class:band-start={band.startsHere}
										class:band-end={band.endsHere}
										style="--start: {band.startCol}; --span: {band.span}; --lane: {band.lane}; background-color: {band.project.color};"
										onclick={() => openProject(band.project)}
										title={band.project.name}
									>
										{#if !band.startsHere}
											<span class="band-continuation-arrow left">‹</span>
										{/if}
										<span class="project-band-title">{band.project.name}</span>
										{#if band.project.isBugBounty}
											<span class="project-band-tag">BB</span>
										{/if}
										{#if !band.endsHere}
											<span class="band-continuation-arrow right">›</span>
										{/if}
									</button>
								{/each}
							</div>
						</div>
					{/each}
				</div>
			</section>
		{/each}

		<div bind:this={bottomSentinel} class="sentinel"></div>
	</div>
</div>

<style>
	.calendar-shell {
		display: flex;
		flex-direction: column;
		overflow: hidden;
	}

	.calendar-toolbar {
		display: flex;
		justify-content: flex-start;
		align-items: center;
		padding: 10px 16px;
		border-bottom: 1px solid var(--tblr-border-color);
		gap: 12px;
		flex-wrap: wrap;
	}

	.group-legend {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
		align-items: center;
	}

	.group-chip {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 3px 10px;
		border-radius: 14px;
		border: 1px solid var(--tblr-border-color);
		background: var(--tblr-body-bg);
		font-size: 0.75rem;
		font-weight: 500;
		color: var(--tblr-body-color);
		cursor: pointer;
		transition: opacity 0.12s ease, background 0.12s ease;
	}

	.group-chip:hover {
		background: rgba(0, 0, 0, 0.04);
	}

	.group-chip.collapsed {
		opacity: 0.45;
		text-decoration: line-through;
	}

	.chip-dot {
		width: 10px;
		height: 10px;
		border-radius: 50%;
		flex-shrink: 0;
	}

	.calendar-scroller {
		min-height: 70vh;
		max-height: 82vh;
		overflow-y: auto;
		overflow-x: hidden;
		background: var(--tblr-body-bg);
		scroll-behavior: auto;
	}

	.sentinel {
		height: 1px;
		width: 100%;
	}

	.month-section {
		padding: 0 0 16px 0;
	}

	.month-section + .month-section {
		border-top: 1px solid var(--tblr-border-color);
	}

	.month-header {
		position: sticky;
		top: 0;
		z-index: 50;
		background: var(--tblr-body-bg);
		padding: 14px 20px 10px;
		border-bottom: 1px solid var(--tblr-border-color);
	}

	.month-title {
		margin: 0;
		font-size: 1.4rem;
		font-weight: 600;
		letter-spacing: -0.02em;
		color: var(--tblr-body-color);
	}

	.month-year {
		font-weight: 400;
		color: var(--tblr-muted);
		margin-left: 8px;
	}

	.weekday-header {
		display: grid;
		grid-template-columns: repeat(7, 1fr);
		padding: 8px 8px 4px;
		font-size: 0.68rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--tblr-muted);
	}

	.weekday-cell {
		padding: 4px 10px;
		text-align: left;
	}

	.weekday-cell.weekend {
		opacity: 0.55;
	}

	.weeks {
		padding: 0 8px;
	}

	.week-row {
		position: relative;
		--band-height: 24px;
		--band-gap: 4px;
		--top-padding: 32px;
	}

	.day-grid {
		display: grid;
		grid-template-columns: repeat(7, 1fr);
		position: relative;
		border-top: 1px solid var(--tblr-border-color);
		min-height: calc(
			var(--top-padding) + var(--lane-count, 1) * (var(--band-height) + var(--band-gap)) + 16px
		);
	}

	.day-cell {
		border-right: 1px solid var(--tblr-border-color);
		padding: 5px 8px;
		min-height: 80px;
		position: relative;
	}

	.day-cell:last-child {
		border-right: none;
	}

	.day-cell.weekend {
		background-color: rgba(var(--tblr-muted-rgb), 0.04);
	}

	.day-cell.out-of-month .day-number {
		color: var(--tblr-muted);
		opacity: 0.45;
	}

	.day-number {
		font-size: 0.82rem;
		font-weight: 500;
		width: 24px;
		height: 24px;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		border-radius: 50%;
		color: var(--tblr-body-color);
		line-height: 1;
	}

	.day-cell.today .day-number {
		background: var(--tblr-primary);
		color: #fff;
		font-weight: 600;
		box-shadow: 0 2px 6px rgba(32, 107, 196, 0.35);
	}

	.project-band-layer {
		position: absolute;
		top: var(--top-padding);
		left: 0;
		right: 0;
		bottom: 4px;
		pointer-events: none;
	}

	.project-band {
		pointer-events: auto;
		position: absolute;
		top: calc(var(--lane) * (var(--band-height) + var(--band-gap)));
		left: calc(var(--start) * (100% / 7));
		width: calc(var(--span) * (100% / 7));
		height: var(--band-height);
		padding: 0 10px;
		margin: 0;
		display: flex;
		align-items: center;
		gap: 6px;
		border: 1px solid rgba(0, 0, 0, 0.08);
		border-radius: 7px;
		color: #fff;
		font-size: 0.8rem;
		font-weight: 600;
		text-align: left;
		cursor: pointer;
		user-select: none;
		overflow: hidden;
		transition: transform 0.12s ease, box-shadow 0.12s ease, filter 0.12s ease;
		box-shadow: 0 1px 2px rgba(0, 0, 0, 0.08);
		container-type: inline-size;
		line-height: 1;
	}

	.project-band.band-start {
		margin-left: 2px;
		width: calc(var(--span) * (100% / 7) - 2px);
	}

	.project-band.band-end {
		width: calc(var(--span) * (100% / 7) - 2px);
	}

	.project-band.band-start.band-end {
		width: calc(var(--span) * (100% / 7) - 4px);
	}

	.project-band:not(.band-start) {
		border-top-left-radius: 0;
		border-bottom-left-radius: 0;
		border-left: none;
	}

	.project-band:not(.band-end) {
		border-top-right-radius: 0;
		border-bottom-right-radius: 0;
		border-right: none;
	}

	.project-band:hover {
		filter: brightness(1.1);
		z-index: 5;
		box-shadow: 0 6px 18px rgba(0, 0, 0, 0.22);
		transform: translateY(-1px);
	}

	.project-band:focus-visible {
		outline: 2px solid #fff;
		outline-offset: -3px;
		z-index: 5;
	}

	.project-band-title {
		flex: 1 1 auto;
		text-overflow: ellipsis;
		overflow: hidden;
		white-space: nowrap;
		min-width: 0;
	}

	.project-band-tag {
		flex: 0 0 auto;
		background: rgba(0, 0, 0, 0.24);
		border-radius: 4px;
		padding: 2px 6px;
		font-size: 0.62rem;
		font-weight: 700;
		letter-spacing: 0.04em;
		white-space: nowrap;
		line-height: 1;
	}

	.band-continuation-arrow {
		flex: 0 0 auto;
		font-size: 0.9rem;
		opacity: 0.7;
		line-height: 1;
	}

	.band-continuation-arrow.left {
		margin-left: -4px;
	}

	.band-continuation-arrow.right {
		margin-right: -4px;
		margin-left: auto;
	}

	@container (max-width: 90px) {
		.project-band-tag { display: none; }
	}

	@container (max-width: 50px) {
		.project-band-title { font-size: 0.7rem; }
		.project-band { padding: 0 5px; gap: 3px; }
	}
</style>
