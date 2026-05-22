<script>
	import { run } from 'svelte/legacy';

	import Avatar from '$lib/components/Avatar.svelte';
	import TaskModal from './TaskModal.svelte';
	import { Fetch } from '$lib/fetchUtil';
	import { onMount } from 'svelte';

	let teams = $state([]);
	let allUsers = [];
	let allTeams = [];
	let editMode = $state(false);
	const { startDate, endDate } = calculateWeek();
	let loading = $state(true);


	// Drag and drop state variables
	let isDragging = false;
	let isResizing = false;
	let resizeDirection = null; // 'left' or 'right'
	let draggedTask = null;
	let dragStartX = 0;
	let dragStartY = 0;
	let dragStartDate = null;
	let dragStartMember = null;
	let dragTargetMember = null; // Add this variable to track the final drop target
	let dragPreviewElement = null;
	let dragPreviewElements = []; // Array to hold all preview elements for linked tasks
	let originalTaskPosition = { dateFrom: null, dateTo: null, member: null };
	let cellWidth = 50; // Default cell width (will be measured on drag start)
	let weekendCellWidth = 22; // Width of weekend cells
	let originalTaskElement = null; // Add a reference to track the original element being dragged
	let originalTaskElements = []; // Array to hold all original elements for linked tasks
	let dragLaneOffset = 0; // Vertical offset of the task within its cell
	let isShiftPressed = false; // Track if shift key is pressed during drag

	// Add state variables to track visible dates
	let visibleStartDate = $state(null);
	let visibleEndDate = $state(null);
	let visibleDays = $state([]);
	let visibleStats = $state({ count: 0, hours: 0 });

	onMount(async () => {
		try {
			// Load all users and teams first
			const usersResponse = await Fetch('/api/profile/all');
			allUsers = usersResponse.users || [];
			allTeams = usersResponse.teams || [];

			// Then try to load user preferences
			const preferences = await Fetch('/api/profile/preferences');

			if (preferences && preferences.swimlaneUsers && preferences.swimlaneUsers.length > 0) {
				// Use saved preferences if available
				teams = preferences.swimlaneUsers;
				console.log('Loaded user preferences for swimlane view');
			} else {
				// Initialize with default data if no preferences exist
				// Try to get current user first
				const profileResponse = await Fetch('/api/profile');

				if (profileResponse && profileResponse.email) {
					// If we have the current user, add them first
					teams = [profileResponse.email];
				} else if (allUsers.length > 0) {
					// Otherwise, use the first few users as default
					teams = allUsers.slice(0, Math.min(5, allUsers.length)).map((user) => user.email);
				}

				// Save these initial preferences
				await savePreferences();
				console.log('Initialized default swimlane users');
			}
		} catch (error) {
			console.error('Error fetching teams or preferences:', error);
		}

		// Initial fetch of calendar events
		await fetchCalendarEvents();

		document.addEventListener('mouseup', handleMouseUp);
		document.addEventListener('mousemove', handleDragMove);
		// Add keyboard event listeners for shift key
		document.addEventListener('keydown', handleKeyDown);
		document.addEventListener('keyup', handleKeyUp);
		// Add scroll event listener
		window.addEventListener('scroll', handleScroll, true);
		// Add resize listener to update visible days when window is resized
		window.addEventListener('resize', updateVisibleDays);

		// Initial calculation of visible days after a small delay
		setTimeout(updateVisibleDays, 100);

		return () => {
			document.removeEventListener('mouseup', handleMouseUp);
			document.removeEventListener('mousemove', handleDragMove);
			document.removeEventListener('keydown', handleKeyDown);
			document.removeEventListener('keyup', handleKeyUp);
			window.removeEventListener('scroll', handleScroll, true);
			window.removeEventListener('resize', updateVisibleDays);
		};
	});

	// Function to save user preferences
	async function savePreferences() {
		try {
			await Fetch('/api/profile/preferences', {
				method: 'PATCH',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					swimlaneUsers: teams
				})
			});
			console.log('Saved swimlane preferences');
		} catch (error) {
			console.error('Error saving preferences:', error);
		}
	}

	// Toggle edit mode for swimlane users
	function toggleEditMode() {
		editMode = !editMode;
		if (!editMode) {
			// Save preferences when exiting edit mode
			savePreferences();
		}
	}

	// Add a user to the swimlane view
	function addUser(email) {
		if (!teams.includes(email)) {
			teams = [...teams, email];
		}
	}

	// Add all members from a team
	function addTeam(teamMembers) {
		const newTeams = [...teams];
		teamMembers.forEach((email) => {
			if (!newTeams.includes(email)) {
				newTeams.push(email);
			}
		});
		teams = newTeams;
	}

	// Remove a user from the swimlane view
	function removeUser(email) {
		teams = teams.filter((member) => member !== email);
	}

	// Filter users for adding to swimlane
	function getAvailableUsers() {
		return allUsers.filter((user) => !teams.includes(user.email));
	}

	// Filter teams that have at least one member not in the current view
	function getAvailableTeams() {
		return allTeams.filter(
			(team) => team.members && team.members.some((member) => !teams.includes(member))
		);
	}

	// Check if a team has members that can be added
	function getAddableTeamMembers(team) {
		return team.members.filter((member) => !teams.includes(member));
	}

	// Check if user is filtered out (when searching)
	let userFilter = $state('');


	async function updateHackers(task, event) {
		if (!event) return;

		// Ensure hackers list has unique emails
		const uniqueHackers = [];
		const emailSet = new Set();

		for (const hacker of event.detail) {
			if (!emailSet.has(hacker.email)) {
				emailSet.add(hacker.email);
				uniqueHackers.push(hacker);
			}
		}

		// Update local state first for immediate UI update
		const taskIndex = calendarEvents.findIndex((e) => e.id === task.id);
		if (taskIndex !== -1) {
			calendarEvents[taskIndex].hackers = uniqueHackers;
			calendarEvents = [...calendarEvents]; // Trigger reactivity
		}

		// Then update on the server
		await Fetch(`/api/planning/${task.id}`, {
			method: 'PATCH',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ hackers: uniqueHackers })
		});
	}

	async function updateStatus(task, event) {
		if (!event) return;

		const status = event.detail.status;

		// Update local state first for immediate UI update
		const taskIndex = calendarEvents.findIndex((e) => e.id === task.id);
		if (taskIndex !== -1) {
			calendarEvents[taskIndex].status = status;
			calendarEvents = [...calendarEvents]; // Trigger reactivity
		}

		// Then update on the server
		await Fetch(`/api/planning/${task.id}`, {
			method: 'PATCH',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ status: status })
		});
	}

	async function updateProjects(task, event) {
		if (!event || !event.detail.projects) return;

		// Ensure projects list has unique IDs
		const uniqueProjects = [];
		const idSet = new Set();

		for (const project of event.detail.projects) {
			if (!idSet.has(project.id)) {
				idSet.add(project.id);
				uniqueProjects.push(project);
			}
		}

		// Update local state first for immediate UI update
		const taskIndex = calendarEvents.findIndex((e) => e.id === task.id);
		if (taskIndex !== -1) {
			calendarEvents[taskIndex].projects = uniqueProjects;
			calendarEvents = [...calendarEvents]; // Trigger reactivity
		}

		// Then update on the server
		await Fetch(`/api/planning/${task.id}`, {
			method: 'PATCH',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ projects: uniqueProjects })
		});
	}

	async function updateTitle(task, event) {
		if (!event) return;
		const title = event.detail.title;

		// Update local state first
		const taskIndex = calendarEvents.findIndex((e) => e.id === task.id);
		if (taskIndex !== -1) {
			calendarEvents[taskIndex].title = title;
			calendarEvents = [...calendarEvents]; // Trigger reactivity
		}

		// Then update on the server
		await Fetch(`/api/planning/${task.id}`, {
			method: 'PATCH',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ title })
		});
	}

	async function updateColor(task, event) {
		if (!event) return;
		const color = event.detail.color;

		// Find the task in calendarEvents array and update its color
		const taskIndex = calendarEvents.findIndex((e) => e.id === task.id);
		if (taskIndex !== -1) {
			calendarEvents[taskIndex].color = color;
			calendarEvents = [...calendarEvents]; // Trigger reactivity

			// Direct DOM update for immediate visual feedback
			const elements = document.querySelectorAll(`.planningid-${task.id}`);
			elements.forEach((element) => {
				element.style.backgroundColor = color;
			});

			await Fetch(`/api/planning/${task.id}`, {
				method: 'PATCH',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ color })
			});
		}
	}

	function calculateWeek() {
		const currentYear = new Date().getFullYear();
		let today = new Date();
		let startDate = new Date(today);
		startDate.setDate(today.getDate() - ((today.getDay() + 6) % 7));
		let endDate = new Date(`${currentYear}-12-31`);
		return { startDate, endDate };
	}

	function generateDays(start, end) {
		const daysList = [];
		for (let d = new Date(start); d <= end; d.setDate(d.getDate() + 1)) {
			const date = new Date(d);
			const isWeekend = date.getDay() === 0 || date.getDay() === 6;
			const dayLabel = isWeekend
				? ''
				: ['SUN', 'MON', 'TUE', 'WED', 'THU', 'FRI', 'SAT'][date.getDay()];

			daysList.push({
				date: date.toISOString().split('T')[0],
				label: dayLabel,
				isWeekend
			});
		}
		return daysList;
	}

	let calendarEvents = $state([]);
	/**
	 * @typedef {Object} Props
	 * @property {boolean} [reload]
	 */

	/** @type {Props} */
	let { reload = false } = $props();
	let days = generateDays(startDate, endDate);

	async function fetchCalendarEvents() {
		try {
			const events = await Fetch(
				`/api/planning?startDate=${formatDateToISO(startDate)}&endDate=${formatDateToISO(endDate)}&pageSize=${1000}`
			);

			// Process all events to add calculated properties
			calendarEvents = (events || []).map((event) => {
				return {
					...event,
					durationDays: calculateDurationDays(event),
					durationHours: calculateDurationHours(event),
					width: calculateTaskWidth(event.dateFrom, event.dateTo),
					color: event.color || '#ae3ec9'
				};
			});
		} catch (error) {
			console.error('Error fetching calendar events:', error);
			calendarEvents = [];
		}
	}

	function calculateDurationDays(event) {
		return (() => {
			// Calculate working days (Mon-Fri only)
			const start = new Date(event.dateFrom);
			const end = new Date(event.dateTo);
			let workdays = 0;

			for (let day = new Date(start); day <= end; day.setDate(day.getDate() + 1)) {
				const dayOfWeek = day.getDay();
				if (dayOfWeek > 0 && dayOfWeek < 6) workdays++;
			}

			return workdays;
		})();
	}

	function calculateDurationHours(event) {
		return (() => {
			// Calculate working hours (7.5h per weekday, no weekends)
			const start = new Date(event.dateFrom);
			const end = new Date(event.dateTo);
			let workdays = 0;

			for (let day = new Date(start); day <= end; day.setDate(day.getDate() + 1)) {
				const dayOfWeek = day.getDay();
				if (dayOfWeek > 0 && dayOfWeek < 6) workdays++;
			}

			// Get number of hackers assigned to the event
			const hackerCount = event.hackers?.length || 1;

			// Multiply hours by number of hackers
			return workdays * 7.5 * hackerCount;
		})();
	}

	function formatDateToISO(date) {
		return new Date(date).toISOString().split('T')[0];
	}


	function calculateTaskWidth(startDate, endDate) {
		// Convert string dates to Date objects if needed
		const start = new Date(startDate);
		const end = new Date(endDate);

		let totalWidth = 0;
		const regularDayWidth = 48; // Adjust based on your regular day width
		const weekendDayWidth = 24; // From your CSS for weekend cells

		for (let day = new Date(start); day <= end; day.setDate(day.getDate() + 1)) {
			const isWeekend = day.getDay() === 0 || day.getDay() === 6;
			totalWidth += isWeekend ? weekendDayWidth : regularDayWidth;
		}

		return totalWidth + 'px';
	}


	// Modal properties
	let showModal = $state(false);
	let selectedTask = $state(null);
	let selectedTaskElement = null;
	let lastClickedTaskId = null;
	let clickPosition = $state({ x: 0, y: 0 });

	// Function to handle task click with improved positioning
	function handleTaskClick(event, task) {
		// Prevent task click when we're ending a drag
		if (isDragging || isResizing) {
			event.stopPropagation();
			return;
		}

		event.stopPropagation();

		// Toggle modal if clicking the same task
		if (showModal && lastClickedTaskId === task.id) {
			closeModal();
			return;
		}

		// Remove selected class from previously selected task if any
		if (selectedTaskElement) {
			selectedTaskElement.classList.remove('task-selected');
		}

		// Set current task as selected
		selectedTask = task;
		lastClickedTaskId = task.id;
		selectedTaskElement = event.currentTarget;
		selectedTaskElement.classList.add('task-selected');

		// Add task-dimmed class to all tasks
		document.querySelectorAll('.task').forEach((taskEl) => {
			if (taskEl !== selectedTaskElement) {
				taskEl.classList.add('task-dimmed');
			}
		});

		// Just capture click position, no need to calculate modal position
		clickPosition = { x: event.clientX, y: event.clientY };

		// Show the modal
		showModal = true;
	}

	// Function to close modal
	function closeModal() {
		showModal = false;
		lastClickedTaskId = null;

		// Remove selected class and dimmed effect from all tasks
		if (selectedTaskElement) {
			selectedTaskElement.classList.remove('task-selected');
			selectedTaskElement = null;
		}

		document.querySelectorAll('.task-dimmed').forEach((taskEl) => {
			taskEl.classList.remove('task-dimmed');
		});
	}

	// Function to open full view in new tab
	function openFullView(taskId) {
		window.open(`/planning/${taskId}/view`, '_blank');
	}

	// Add selection tracking variables
	let isSelecting = false;
	let selectionStart = null;
	let selectedCells = [];

	// Start selection on mouse down
	function handleMouseDown(e, day, member) {
		// Don't start selection if clicking on a task
		if (e.target.closest('.task')) return;

		isSelecting = true;
		selectionStart = { day, member };
		selectedCells = [{ day, member }];

		// Add selection class to the first cell
		e.currentTarget.classList.add('cell-selecting');
	}

	// Modify the handleMouseMove function to allow bidirectional selection
	function handleMouseMove(e, day, member) {
		if (!isSelecting) return;

		// Only allow selection in the same row (same member)
		if (member !== selectionStart.member) return;

		// Clear all selection styling first
		document.querySelectorAll('.cell-selecting').forEach((cell) => {
			cell.classList.remove('cell-selecting');
		});

		// Get start and current dates for comparison
		const startDate = new Date(selectionStart.day.date);
		const currentDate = new Date(day.date);

		// Determine the date range (works in both directions)
		const minDate = startDate < currentDate ? startDate : currentDate;
		const maxDate = startDate > currentDate ? startDate : currentDate;

		// Rebuild the selectedCells array with all cells in the range
		selectedCells = [];

		// Find all cells in the range and mark them as selected
		days.forEach((dayItem) => {
			const dayDate = new Date(dayItem.date);
			if (dayDate >= minDate && dayDate <= maxDate) {
				// Add to selected cells array
				selectedCells.push({ day: dayItem, member });

				// Find and style the cell
				const cellSelector = `td[data-date="${dayItem.date}"][data-member="${member}"]`;
				const cell = document.querySelector(cellSelector);
				if (cell) {
					cell.classList.add('cell-selecting');
				}
			}
		});
	}

	// Function to update which days are visible in the viewport - improved
	function updateVisibleDays() {
		const tableContainer = document.querySelector('.table-responsive');
		if (!tableContainer) return;

		const tableRect = tableContainer.getBoundingClientRect();
		const headerCells = document.querySelectorAll('thead th:not(.first-col)');

		// Find first and last visible date cells
		let firstVisibleIndex = -1;
		let lastVisibleIndex = -1;

		headerCells.forEach((cell, index) => {
			const cellRect = cell.getBoundingClientRect();
			// Check if cell is at least partially visible
			if (cellRect.left < tableRect.right && cellRect.right > tableRect.left) {
				if (firstVisibleIndex === -1) firstVisibleIndex = index;
				lastVisibleIndex = index;
			}
		});

		// Update visible dates with new arrays to trigger reactivity
		if (firstVisibleIndex !== -1 && lastVisibleIndex !== -1) {
			// Use a new Date string to ensure reactivity
			visibleStartDate = days[firstVisibleIndex]?.date ? days[firstVisibleIndex].date : null;
			visibleEndDate = days[lastVisibleIndex]?.date ? days[lastVisibleIndex].date : null;
			visibleDays = [...days.slice(firstVisibleIndex, lastVisibleIndex + 1)];
		}
	}

	// Helper function to get all task elements for a given task ID
	function getAllTaskElements(taskId) {
		return Array.from(document.querySelectorAll(`.planningid-${taskId}`));
	}

	// Helper function to get all members (hackers) assigned to a task
	function getAllTaskMembers(task) {
		return task.hackers?.map((h) => h.email) || [];
	}

	// Start dragging a task
	function handleTaskDragStart(event, task, member) {
		event.stopPropagation();

		// Don't start drag if we're in edit mode or task is being resized
		if (editMode || isResizing) return;

		isDragging = true;
		draggedTask = { ...task };
		dragStartX = event.clientX;
		dragStartY = event.clientY;

		// Store original position for calculations
		originalTaskPosition = {
			dateFrom: task.dateFrom,
			dateTo: task.dateTo,
			member: member // Just use the first hacker for simplicity
		};

		// Store the current date and member for calculations
		const taskStartDate = new Date(task.dateFrom).toISOString().split('T')[0];
		dragStartDate = taskStartDate;
		dragStartMember = originalTaskPosition.member;

		// Measure the cell widths for accurate positioning
		const regularCell = document.querySelector('td:not(.weekend):not(.first-col)');
		const weekendCell = document.querySelector('td.weekend');
		if (regularCell) cellWidth = regularCell.offsetWidth;
		if (weekendCell) weekendCellWidth = weekendCell.offsetWidth;

		// Get ALL instances of this task (for all assigned hackers)
		originalTaskElements = getAllTaskElements(task.id);
		originalTaskElement = event.currentTarget; // Keep track of the one clicked
		dragLaneOffset = parseFloat(originalTaskElement.style.top) || 0;

		// Hide all instances of the task
		originalTaskElements.forEach((taskEl) => {
			taskEl.style.visibility = 'hidden';
		});

		// If shift is already pressed, show all original task instances with FULL opacity
		if (event.shiftKey) {
			originalTaskElements.forEach((taskEl) => {
				taskEl.style.visibility = 'visible';
				taskEl.style.opacity = '1'; // Full opacity instead of dimmed
			});
			isShiftPressed = true;
		}

		// Dim all other tasks during drag (but not the instances of the same task)
		document.querySelectorAll('.task').forEach((taskEl) => {
			if (!originalTaskElements.includes(taskEl)) {
				taskEl.classList.add('task-dimmed-drag');
			}
		});

		// Create and position preview elements for ALL instances
		createDragPreviewsForLinkedTasks(task);

		// Add global class to indicate dragging state
		document.body.classList.add('dragging-active');
	}

	// Create preview elements for all linked task instances
	function createDragPreviewsForLinkedTasks(task) {
		// Remove any existing previews
		dragPreviewElements.forEach((el) => el.remove());
		dragPreviewElements = [];

		const taskMembers = getAllTaskMembers(task);
		const allElements = getAllTaskElements(task.id);

		// Create a preview for each instance of the task
		allElements.forEach((sourceElement) => {
			const preview = sourceElement.cloneNode(true);
			preview.classList.add('task-preview');
			preview.style.position = 'fixed';
			preview.style.opacity = '1'; // Full opacity
			preview.style.pointerEvents = 'none';
			preview.style.zIndex = '1000';
			preview.style.visibility = 'visible';

			// Position it at the same place as the original
			const rect = sourceElement.getBoundingClientRect();
			preview.style.width = rect.width + 'px';
			preview.style.height = rect.height + 'px';
			preview.style.top = rect.top + 'px';
			preview.style.left = rect.left + 'px';

			// Store the member email for this preview
			const memberAttr = sourceElement.closest('td')?.getAttribute('data-member');
			if (memberAttr) {
				preview.dataset.memberEmail = memberAttr;
				preview.dataset.laneOffset = parseFloat(sourceElement.style.top) || 0;
			}

			document.body.appendChild(preview);
			dragPreviewElements.push(preview);
		});
	}

	// Create a visual preview element for dragging - Improved visibility
	function createDragPreview(sourceElement, task) {
		// Remove any existing preview
		if (dragPreviewElement) {
			dragPreviewElement.remove();
		}

		// Clone the task element for the preview
		dragPreviewElement = sourceElement.cloneNode(true);
		dragPreviewElement.classList.add('task-preview');
		dragPreviewElement.style.position = 'fixed'; // Change to fixed positioning
		dragPreviewElement.style.opacity = '1'; // Full opacity - not transparent
		dragPreviewElement.style.pointerEvents = 'none';
		dragPreviewElement.style.zIndex = '1000';
		dragPreviewElement.style.visibility = 'visible'; // Explicitly set to visible

		// Position it at the same place as the original
		const rect = sourceElement.getBoundingClientRect();
		dragPreviewElement.style.width = rect.width + 'px';
		dragPreviewElement.style.height = rect.height + 'px';
		dragPreviewElement.style.top = rect.top + 'px';
		dragPreviewElement.style.left = rect.left + 'px';

		// Add to body
		document.body.appendChild(dragPreviewElement);

		// Extra check - make sure preview is visible after a brief delay
		setTimeout(() => {
			if (dragPreviewElement) {
				dragPreviewElement.style.visibility = 'visible';
			}
		}, 10);
	}

	// Handle mouse move during drag - Fixed visibility issues
	function handleDragMove(event) {
		if (!isDragging && !isResizing) return;

		// Check if shift key is pressed directly from the event
		const wasShiftPressed = isShiftPressed;
		isShiftPressed = event.shiftKey;

		// Toggle original task visibility if shift state changed
		if (originalTaskElements.length > 0 && wasShiftPressed !== isShiftPressed) {
			originalTaskElements.forEach((taskEl) => {
				if (isShiftPressed) {
					taskEl.style.visibility = 'visible';
					taskEl.style.opacity = '1'; // Full opacity
				} else {
					taskEl.style.visibility = 'hidden';
				}
			});
		}

		if (isResizing) {
			handleResizeMove(event);
			return;
		}

		if (dragPreviewElements.length === 0 || !draggedTask) return;

		// Disable pointer events on all tasks once actual dragging starts
		document.querySelectorAll('.task').forEach((el) => {
			el.style.pointerEvents = 'none';
		});

		// Find the cell under the cursor for snapping
		// Temporarily hide all previews for hit testing
		dragPreviewElements.forEach((preview) => {
			preview.style.opacity = '0';
		});

		const elemBelow = document.elementFromPoint(event.clientX, event.clientY);

		// Restore opacity immediately
		dragPreviewElements.forEach((preview) => {
			preview.style.opacity = '1';
		});

		const cell = elemBelow?.closest('td:not(.first-col)');

		if (cell) {
			const cellDate = cell.getAttribute('data-date');
			const cellMember = cell.getAttribute('data-member');

			if (cellDate && cellMember) {
				// Store the current target member for use when the drag ends
				dragTargetMember = cellMember;

				// Calculate new dates
				const daysDiff = calculateDaysDifference(dragStartDate, cellDate);

				const newStartDate = new Date(originalTaskPosition.dateFrom);
				newStartDate.setDate(newStartDate.getDate() + daysDiff);

				const newEndDate = new Date(originalTaskPosition.dateTo);
				newEndDate.setDate(newEndDate.getDate() + daysDiff);

				// Update the preview task dates
				draggedTask.dateFrom = newStartDate.toISOString().split('T')[0];
				draggedTask.dateTo = newEndDate.toISOString().split('T')[0];

				// Position all preview elements at their appropriate cells
				dragPreviewElements.forEach((preview) => {
					const previewMember = preview.dataset.memberEmail;
					const laneOffset = parseFloat(preview.dataset.laneOffset) || 5;

					if (previewMember) {
						positionDragPreviewAtCellForMember(
							preview,
							draggedTask.dateFrom,
							previewMember,
							laneOffset,
							draggedTask
						);
					}
				});
			}
		} else {
			// If not over a cell, just move all previews with the cursor
			const deltaX = event.clientX - dragStartX;
			const deltaY = event.clientY - dragStartY;

			dragPreviewElements.forEach((preview) => {
				const rect = preview.getBoundingClientRect();
				preview.style.top = rect.top + deltaY + 'px';
				preview.style.left = rect.left + deltaX + 'px';
			});

			// Reset start positions for relative movement
			dragStartX = event.clientX;
			dragStartY = event.clientY;
		}
	}

	// Calculate days difference between two date strings
	function calculateDaysDifference(date1, date2) {
		const d1 = new Date(date1);
		const d2 = new Date(date2);
		return Math.round((d2 - d1) / (1000 * 60 * 60 * 24));
	}

	// Position the drag preview at the specified date and member cell
	function positionDragPreviewAtCell(dateStr, memberEmail) {
		const cellSelector = `td[data-date="${dateStr}"][data-member="${memberEmail}"]`;
		const targetCell = document.querySelector(cellSelector);

		if (targetCell) {
			const cellRect = targetCell.getBoundingClientRect();
			const taskDuration = calculateTaskWidth(draggedTask.dateFrom, draggedTask.dateTo);

			// Use lane offset only when staying on the same member, otherwise top of row
			const offset = memberEmail === dragStartMember ? dragLaneOffset : 5;
			dragPreviewElement.style.top = cellRect.top + offset + 'px';
			dragPreviewElement.style.left = cellRect.left + 'px';
			dragPreviewElement.style.width = taskDuration;
			dragPreviewElement.style.visibility = 'visible';
		}
	}

	// Position a specific preview element at the specified date and member cell
	function positionDragPreviewAtCellForMember(
		previewElement,
		dateStr,
		memberEmail,
		laneOffset,
		task
	) {
		const cellSelector = `td[data-date="${dateStr}"][data-member="${memberEmail}"]`;
		const targetCell = document.querySelector(cellSelector);

		if (targetCell) {
			const cellRect = targetCell.getBoundingClientRect();
			const taskDuration = calculateTaskWidth(task.dateFrom, task.dateTo);

			previewElement.style.top = cellRect.top + laneOffset + 'px';
			previewElement.style.left = cellRect.left + 'px';
			previewElement.style.width = taskDuration;
			previewElement.style.visibility = 'visible';
		}
	}

	// Start resizing a task
	function handleResizeStart(event, task, member, direction) {
		event.stopPropagation();
		if (editMode) return;

		isResizing = true;
		resizeDirection = direction;
		draggedTask = { ...task };
		dragStartX = event.clientX;

		// Store original position
		originalTaskPosition = {
			dateFrom: task.dateFrom,
			dateTo: task.dateTo,
			member: member
		};

		// Get ALL instances of this task (for all assigned hackers)
		originalTaskElements = getAllTaskElements(task.id);
		originalTaskElement = event.target.closest('.task');
		dragLaneOffset = parseFloat(originalTaskElement.style.top) || 0;

		// Store the original preview positions for all instances
		originalTaskElements.forEach((taskEl) => {
			const rect = taskEl.getBoundingClientRect();
			taskEl._previewTop = rect.top;
			taskEl._previewLeft = rect.left;
			taskEl.style.visibility = 'hidden';
		});

		// Dim all other tasks during resize (but not the instances of the same task)
		document.querySelectorAll('.task').forEach((taskEl) => {
			if (!originalTaskElements.includes(taskEl)) {
				taskEl.classList.add('task-dimmed-drag');
			}
		});

		// Create and position previews for ALL instances
		createResizePreviewsForLinkedTasks(task);

		document.body.classList.add('resizing-active');
	}

	// Create preview elements for all linked task instances during resize
	function createResizePreviewsForLinkedTasks(task) {
		// Remove any existing previews
		dragPreviewElements.forEach((el) => el.remove());
		dragPreviewElements = [];

		const allElements = getAllTaskElements(task.id);

		// Create a preview for each instance of the task
		allElements.forEach((sourceElement) => {
			const preview = sourceElement.cloneNode(true);
			preview.classList.add('task-preview');
			preview.style.position = 'fixed';
			preview.style.opacity = '1'; // Full opacity
			preview.style.pointerEvents = 'none';
			preview.style.zIndex = '1000';
			preview.style.visibility = 'visible';

			// Position it at the same place as the original
			const rect = sourceElement.getBoundingClientRect();
			preview.style.width = rect.width + 'px';
			preview.style.height = rect.height + 'px';
			preview.style.top = rect.top + 'px';
			preview.style.left = rect.left + 'px';

			// Store the member email and original position for this preview
			const memberAttr = sourceElement.closest('td')?.getAttribute('data-member');
			if (memberAttr) {
				preview.dataset.memberEmail = memberAttr;
				preview.dataset.previewTop = rect.top;
				preview.dataset.previewLeft = rect.left;
			}

			document.body.appendChild(preview);
			dragPreviewElements.push(preview);
		});
	}

	// Handle resizing movement - Keep task in original lane position
	function handleResizeMove(event) {
		if (!isResizing || dragPreviewElements.length === 0) return;

		// Disable pointer events on all tasks to allow hit testing through them
		document.querySelectorAll('.task').forEach((el) => {
			el.style.pointerEvents = 'none';
		});

		// Hide all previews temporarily to detect cells underneath
		dragPreviewElements.forEach((preview) => {
			preview.style.opacity = '0';
		});

		const elemBelow = document.elementFromPoint(event.clientX, event.clientY);

		// Restore opacity immediately
		dragPreviewElements.forEach((preview) => {
			preview.style.opacity = '1';
		});

		const cell = elemBelow?.closest('td:not(.first-col)');

		if (cell) {
			const cellDate = cell.getAttribute('data-date');
			const cellMember = cell.getAttribute('data-member');

			// Allow resize in any member lane (not just original)
			if (cellDate) {
				// Store the proposed dates temporarily
				let newDateFrom = draggedTask.dateFrom;
				let newDateTo = draggedTask.dateTo;

				// Calculate new dates based on resize direction
				if (resizeDirection === 'left') {
					newDateFrom = cellDate;
				} else if (resizeDirection === 'right') {
					newDateTo = cellDate;
				}

				// Ensure end date isn't before start date
				const startDate = new Date(newDateFrom);
				const endDate = new Date(newDateTo);

				if (endDate < startDate) {
					if (resizeDirection === 'left') {
						newDateFrom = newDateTo;
					} else {
						newDateTo = newDateFrom;
					}
				}

				// Update draggedTask dates for when drag ends
				draggedTask.dateFrom = newDateFrom;
				draggedTask.dateTo = newDateTo;

				// Calculate new width
				const newWidth = calculateTaskWidth(draggedTask.dateFrom, draggedTask.dateTo);

				// Update all preview elements
				dragPreviewElements.forEach((preview) => {
					preview.style.width = newWidth;

					if (resizeDirection === 'left') {
						// When resizing from left, update the left position but keep the same lane (top position)
						const previewMember = preview.dataset.memberEmail;
						const cellSelector = `td[data-date="${draggedTask.dateFrom}"][data-member="${previewMember}"]`;
						const targetCell = document.querySelector(cellSelector);

						if (targetCell) {
							const cellRect = targetCell.getBoundingClientRect();
							const originalTop = preview.dataset.previewTop;
							// Keep the original top position (lane), only update left
							preview.style.left = cellRect.left + 'px';
							preview.style.top = originalTop + 'px';
						}
					}
					// For right resize, position stays the same, only width changes
				});
			}
		}

		// Always ensure all previews are visible after positioning
		dragPreviewElements.forEach((preview) => {
			preview.style.visibility = 'visible';
		});

		// Reset start X for next calculation
		dragStartX = event.clientX;
	}

	// Handle end of drag or resize operation
	async function handleTaskDragEnd() {
		if (!isDragging && !isResizing) return;

		try {
			if (draggedTask) {
				// Check if the dates changed or if the drop target is different from the start member
				const datesChanged =
					draggedTask.dateFrom !== originalTaskPosition.dateFrom ||
					draggedTask.dateTo !== originalTaskPosition.dateTo;

				const memberChanged = dragTargetMember && dragTargetMember !== dragStartMember;

				const hasChanges = datesChanged || memberChanged;

				if (hasChanges) {
					// Get the current hackers
					let updatedHackers = [...draggedTask.hackers];

					if (memberChanged) {
						if (isShiftPressed) {
							// When shift is pressed, add the target member without removing the original
							// Check if target member is already in the list to avoid duplicates
							if (!updatedHackers.some((h) => h.email === dragTargetMember)) {
								updatedHackers.push({ email: dragTargetMember });
							}
						} else {
							// Default behavior (replace): Remove the original member
							updatedHackers = updatedHackers.filter((h) => h.email !== dragStartMember);

							// Check if target member is already in the list to avoid duplicates
							if (!updatedHackers.some((h) => h.email === dragTargetMember)) {
								updatedHackers.push({ email: dragTargetMember });
							}
						}
					}

					// Additional uniqueness check - create a clean list with no duplicates
					const uniqueEmails = new Set();
					updatedHackers = updatedHackers.filter((h) => {
						if (uniqueEmails.has(h.email)) {
							return false; // Skip duplicates
						}
						uniqueEmails.add(h.email);
						return true;
					});

					// Prepare data for the update
					const updateData = {
						dateFrom: draggedTask.dateFrom,
						dateTo: draggedTask.dateTo,
						hackers: updatedHackers
					};

					// Update the server
					await Fetch(`/api/planning/${draggedTask.id}`, {
						method: 'PATCH',
						headers: {
							'Content-Type': 'application/json'
						},
						body: JSON.stringify(updateData)
					});

					// Update local state
					const taskIndex = calendarEvents.findIndex((e) => e.id === draggedTask.id);
					if (taskIndex !== -1) {
						calendarEvents[taskIndex] = {
							...calendarEvents[taskIndex],
							...updateData,
							durationDays: calculateDurationDays(updateData),
							durationHours: calculateDurationHours(updateData),
							width: calculateTaskWidth(updateData.dateFrom, updateData.dateTo)
						};
						calendarEvents = [...calendarEvents]; // Trigger reactivity
					}
				}
			}
		} catch (error) {
			console.error('Error updating task:', error);
			// Refresh calendar events in case of error
			await fetchCalendarEvents();
		} finally {
			// Restore visibility of all original elements
			if (originalTaskElements.length > 0) {
				originalTaskElements.forEach((taskEl) => {
					taskEl.style.visibility = '';
				});
				originalTaskElements = [];
				originalTaskElement = null;
			}

			// Clean up all preview elements
			if (dragPreviewElements.length > 0) {
				dragPreviewElements.forEach((preview) => preview.remove());
				dragPreviewElements = [];
			}

			// Fallback cleanup for old single preview element
			if (dragPreviewElement) {
				dragPreviewElement.remove();
				dragPreviewElement = null;
			}

			// Remove dimming from all tasks
			document.querySelectorAll('.task-dimmed-drag').forEach((taskEl) => {
				taskEl.classList.remove('task-dimmed-drag');
			});

			isDragging = false;
			isResizing = false;
			draggedTask = null;
			dragTargetMember = null; // Reset the target member
			document.querySelectorAll('.task').forEach((el) => {
				el.style.pointerEvents = '';
			});
			document.body.classList.remove('dragging-active', 'resizing-active');
		}
	}

	// End selection on mouse up
	async function handleMouseUp() {
		// Handle existing selection logic
		if (isSelecting) {
			if (selectedCells.length === 1) {
				clearSelection();
				return;
			}

			// Create new task
			try {
				const response = await Fetch(`/api/planning/new`, {
					method: 'POST',
					headers: {
						'Content-Type': 'application/json'
					},
					body: JSON.stringify({
						title: 'New Task',
						projects: [],
						dateFrom: selectedCells[0].day.date,
						dateTo: selectedCells[selectedCells.length - 1].day.date,
						hackers: [...new Set(selectedCells.map((cell) => cell.member))].map((email) => ({
							email
						}))
					})
				});

				// If successful, add the new task to local state
				if (response && response.id) {
					const newTask = {
						...response,
						durationDays: calculateDurationDays(response),
						durationHours: calculateDurationHours(response),
						width: calculateTaskWidth(response.dateFrom, response.dateTo),
						color: response.color || '#ae3ec9'
					};

					calendarEvents = [...calendarEvents, newTask];
				} else {
					// If no valid response, fetch all events again
					await fetchCalendarEvents();
				}
			} catch (error) {
				console.error('Error creating new task:', error);
				await fetchCalendarEvents();
			}

			// Clear selection after a short delay
			setTimeout(clearSelection, 500);
		}

		// Handle drag end
		await handleTaskDragEnd();
	}

	// Clear the visual selection
	function clearSelection() {
		isSelecting = false;
		document.querySelectorAll('.cell-selecting').forEach((cell) => {
			cell.classList.remove('cell-selecting');
		});
		selectedCells = [];
	}

	// Function to filter events for a specific day and member
	function computeMemberLanes(memberEmail) {
		if (!calendarEvents?.length) return { laneCount: 1, laneMap: {}, hasOverlaps: {} };

		// Keep original array order to preserve visual lane stability
		const memberEvents = calendarEvents.filter((event) =>
			event.hackers?.some((h) => h.email === memberEmail)
		);

		if (memberEvents.length === 0) return { laneCount: 1, laneMap: {}, hasOverlaps: {} };

		const lanes = []; // each lane tracks occupied intervals
		const laneMap = {};
		const hasOverlaps = {}; // Track which tasks have overlaps

		for (const event of memberEvents) {
			const eventStart = new Date(event.dateFrom);
			const eventEnd = new Date(event.dateTo);
			let assigned = false;

			// Try to fit in existing lanes (first one without overlap)
			for (let i = 0; i < lanes.length; i++) {
				const hasOverlap = lanes[i].some(
					(interval) => eventStart <= interval.end && eventEnd >= interval.start
				);
				if (!hasOverlap) {
					lanes[i].push({ start: eventStart, end: eventEnd });
					laneMap[event.id] = i;
					assigned = true;
					break;
				}
			}

			if (!assigned) {
				laneMap[event.id] = lanes.length;
				lanes.push([{ start: eventStart, end: eventEnd }]);
			}
		}

		// Now determine which tasks have overlaps
		for (const event of memberEvents) {
			const eventStart = new Date(event.dateFrom);
			const eventEnd = new Date(event.dateTo);

			// Check if this event overlaps with any other event
			const overlapsWithOther = memberEvents.some((otherEvent) => {
				if (otherEvent.id === event.id) return false;
				const otherStart = new Date(otherEvent.dateFrom);
				const otherEnd = new Date(otherEvent.dateTo);
				return eventStart <= otherEnd && eventEnd >= otherStart;
			});

			hasOverlaps[event.id] = overlapsWithOther;
		}

		return { laneCount: Math.max(1, lanes.length), laneMap, hasOverlaps };
	}


	function getEventForDayMemberLane(dayDate, memberEmail, laneIndex, laneMap) {
		if (!calendarEvents?.length) return null;

		const isFirstDay = formatDateToISO(startDate) === dayDate;

		const matchingEvent = calendarEvents.find((event) => {
			if (laneMap[event.id] !== laneIndex) return false;
			if (!event.hackers.some((h) => h.email === memberEmail)) return false;

			const eventStartDate = new Date(event.dateFrom).toISOString().split('T')[0];
			return isFirstDay ? eventStartDate <= dayDate : eventStartDate === dayDate;
		});

		return matchingEvent || null;
	}

	// Keep the old function for drag/drop compatibility
	function getEventsForDayAndMember(dayDate, memberEmail) {
		if (!calendarEvents?.length) return null;

		if (formatDateToISO(startDate) === dayDate) {
			const matchingEvent = calendarEvents.find((event) => {
				const eventStartDate = new Date(event.dateFrom).toISOString().split('T')[0];
				return eventStartDate <= dayDate && event.hackers.some((h) => h.email === memberEmail);
			});
			return matchingEvent || null;
		}

		const matchingEvent = calendarEvents.find((event) => {
			const eventStartDate = new Date(event.dateFrom).toISOString().split('T')[0];
			return eventStartDate === dayDate && event.hackers.some((h) => h.email === memberEmail);
		});

		return matchingEvent || null;
	}

	// Improved scroll handler with debounce for better performance
	let scrollTimeout;
	function handleScroll() {
		if (isDragging || isResizing) {
			// Handle drag preview positioning during drag/resize operations
			if (dragPreviewElements.length === 0 || !draggedTask) return;

			if (isDragging && dragTargetMember) {
				// Update all preview elements during drag
				dragPreviewElements.forEach((preview) => {
					const previewMember = preview.dataset.memberEmail;
					const laneOffset = parseFloat(preview.dataset.laneOffset) || 5;

					if (previewMember) {
						positionDragPreviewAtCellForMember(
							preview,
							draggedTask.dateFrom,
							previewMember,
							laneOffset,
							draggedTask
						);
					}
				});
			} else if (isResizing) {
				// For resizing, keep stable in original lane - only update horizontal position
				const newWidth = calculateTaskWidth(draggedTask.dateFrom, draggedTask.dateTo);

				dragPreviewElements.forEach((preview) => {
					const previewMember = preview.dataset.memberEmail;
					const cellSelector = `td[data-date="${draggedTask.dateFrom}"][data-member="${previewMember}"]`;
					const targetCell = document.querySelector(cellSelector);

					if (targetCell) {
						const cellRect = targetCell.getBoundingClientRect();
						const originalTop = preview.dataset.previewTop;

						// Keep the original top position (lane), update left and width
						preview.style.left = cellRect.left + 'px';
						preview.style.top = originalTop + 'px';
						preview.style.width = newWidth;
					}
				});
			}
		} else {
			// Debounce the visible days calculation for better performance
			clearTimeout(scrollTimeout);
			scrollTimeout = setTimeout(() => {
				updateVisibleDays();
			}, 100);
		}
	}

	// Track shift key state
	function handleKeyDown(event) {
		if (event.key === 'Shift') {
			isShiftPressed = true;
			// Show all original task instances when shift is pressed (with FULL opacity)
			if (originalTaskElements.length > 0) {
				originalTaskElements.forEach((taskEl) => {
					taskEl.style.visibility = 'visible';
					taskEl.style.opacity = '1'; // Full opacity instead of dimmed
				});
			}
		}
	}

	function handleKeyUp(event) {
		if (event.key === 'Shift') {
			isShiftPressed = false;
			// Hide all original task instances when shift is released
			if (originalTaskElements.length > 0 && isDragging) {
				originalTaskElements.forEach((taskEl) => {
					taskEl.style.visibility = 'hidden';
				});
			}
		}
	}
	// Change the loading condition to only check if teams are loaded
	run(() => {
		loading = !teams?.length;
	});
	let filteredAvailableUsers = $derived(getAvailableUsers().filter(
		(user) =>
			user.email.toLowerCase().includes(userFilter.toLowerCase()) ||
			(user.name && user.name.toLowerCase().includes(userFilter.toLowerCase()))
	));
	let filteredAvailableTeams = $derived(getAvailableTeams().filter(
		(team) =>
			team.name.toLowerCase().includes(userFilter.toLowerCase()) ||
			team.members.some((email) => email.toLowerCase().includes(userFilter.toLowerCase()))
	));
	run(() => {
		if (!reload) {
			fetchCalendarEvents().then(() => {
				// Update visible days after calendar events are loaded
				setTimeout(updateVisibleDays, 100);
			});
		}
	});
	// Create a reactive declaration for visible stats that updates whenever
	// visibleStartDate, visibleEndDate, or calendarEvents change
	run(() => {
		if (visibleStartDate && visibleEndDate && calendarEvents.length) {
			const start = new Date(visibleStartDate);
			const end = new Date(visibleEndDate);

			// Filter events that overlap with visible date range AND have at least one hacker assigned
			const visibleEvents = calendarEvents.filter((event) => {
				const eventStart = new Date(event.dateFrom);
				const eventEnd = new Date(event.dateTo);
				// Check if event overlaps with visible range and has at least one hacker
				return eventStart <= end && eventEnd >= start && event.hackers && event.hackers.length > 0;
			});

			// Calculate total hours for visible events
			const totalHours = visibleEvents.reduce((total, event) => {
				return total + (event.durationHours || 0);
			}, 0);

			visibleStats = {
				count: visibleEvents.length,
				hours: totalHours
			};
		} else {
			visibleStats = { count: 0, hours: 0 };
		}
	});
	// Reactive lane computation per member (depends on both teams and calendarEvents)
	let memberLanes = $derived((() => {
		const _events = calendarEvents; // explicit dependency
		const result = {};
		for (const member of teams) {
			result[member] = computeMemberLanes(member);
		}
		return result;
	})());
</script>

{#if loading}
	<div class="card">
		<div class="card-body text-center">
			<div class="spinner-border text-primary" role="status">
				<span class="visually-hidden">Loading...</span>
			</div>
		</div>
	</div>
{:else}
	<div class="card">
		<div class="card-header d-flex justify-content-between align-items-center">
			<h3 class="card-title">Swimlane View</h3>
			<button
				class="btn btn-sm {editMode ? 'btn-primary' : 'btn-outline-primary'}"
				onclick={toggleEditMode}
			>
				{editMode ? 'Save Layout' : 'Customize View'}
			</button>
		</div>

		{#if editMode}
			<div class="card-body">
				<div class="row">
					<div class="col-md-6">
						<h4>Current Users</h4>
						<div class="avatar-list mb-3">
							{#each teams as member (member)}
								<div class="avatar-container">
									<Avatar
										email={member}
										option={{
											showName: true,
											size: 'sm',
											emptyFields: false,
											circle: true,
											tooltipEnabled: false
										}}
									/>
									<i class="overlay ti ti-x rounded-circle" onclick={() => removeUser(member)}></i>
								</div>
							{/each}
						</div>
					</div>
					<div class="col-md-6">
						<h4>Add Users or Teams</h4>
						<div class="mb-3">
							<input
								type="text"
								class="form-control"
								placeholder="Filter users or teams..."
								bind:value={userFilter}
							/>
						</div>

						<div class="users-list-container">
							{#if filteredAvailableTeams.length > 0}
								<h5>Teams</h5>
								<ul class="list-group mb-3">
									{#each filteredAvailableTeams as team}
										<li class="list-group-item d-flex justify-content-between align-items-center">
											<span>{team.name}</span>
											<button
												class="btn btn-sm btn-outline-primary"
												onclick={() => addTeam(getAddableTeamMembers(team))}
											>
												Add Available Members
											</button>
										</li>
									{/each}
								</ul>
							{/if}

							{#if filteredAvailableUsers.length > 0}
								<h5>Users</h5>
								<ul class="list-group">
									{#each filteredAvailableUsers as user}
										<li
											class="list-group-item d-flex justify-content-between align-items-center cursor-pointer"
											onclick={() => addUser(user.email)}
										>
											<Avatar
												email={user.email}
												option={{
													showName: true,
													size: 'sm',
													emptyFields: false,
													circle: true,
													tooltipEnabled: false
												}}
											/>
											<i class="ti ti-plus"></i>
										</li>
									{/each}
								</ul>
							{/if}

							{#if filteredAvailableUsers.length === 0 && filteredAvailableTeams.length === 0}
								<div class="alert alert-info">No more users or teams available to add</div>
							{/if}
						</div>
					</div>
				</div>
			</div>
		{/if}

		<div class="table-responsive small">
			<table class="table card-table">
				<thead>
					<tr>
						<th class="sticky-col first-col"></th>
						{#each days as day}
							<th class:weekend={day.isWeekend}>
								<div>{new Date(day.date).toLocaleString('default', { month: 'short' })}</div>
								<div>{new Date(day.date).getDate()}</div>
								<div>{day.label}</div>
							</th>
						{/each}
					</tr>
				</thead>
				<tbody>
					{#each teams as member, index}
						{@const lanes = memberLanes[member] || { laneCount: 1, laneMap: {}, hasOverlaps: {} }}
						{@const laneHeight = Math.max(16, Math.floor(55 / lanes.laneCount))}
						{@const rowHeight = laneHeight * lanes.laneCount + 10}
						{@const compact = lanes.laneCount > 2}
						<tr>
							<td class="sticky-col first-col text-muted" style="min-width:20em">
								<Avatar
									email={member}
									option={{
										showName: true,
										size: 'md',
										emptyFields: false,
										circle: true,
										tooltipEnabled: false
									}}
								/>
							</td>
							{#each days as day}
								<td
									class="cell"
									class:weekend={day.isWeekend}
									style="height: {rowHeight}px"
									onmousedown={(e) => handleMouseDown(e, day, member)}
									onmousemove={(e) => handleMouseMove(e, day, member)}
									ondragstart={(e) => e.preventDefault()}
									data-date={day.date}
									data-member={member}
								>
									{#if calendarEvents?.length}
										{#each Array(lanes.laneCount) as _, laneIdx}
											{@const calendar = getEventForDayMemberLane(
												day.date,
												member,
												laneIdx,
												lanes.laneMap
											)}
											{#if calendar}
												{@const taskHasOverlap = lanes.hasOverlaps[calendar.id]}
												{@const taskHeight = taskHasOverlap ? laneHeight - 2 : rowHeight - 12}
												{@const taskTop = taskHasOverlap ? 5 + laneIdx * laneHeight : 5}
												<!-- svelte-ignore a11y_click_events_have_key_events -->
												<!-- svelte-ignore a11y_no_static_element_interactions -->
												<div
													class:task-done={calendar.status === 'done'}
													class="task planningid-{calendar.id}"
													class:task-compact={compact && taskHasOverlap}
													onclick={(e) => handleTaskClick(e, calendar)}
													onmousedown={(e) => handleTaskDragStart(e, calendar, member)}
													style="width: {calendar.width}; background-color: {calendar.color}; top: {taskTop}px; height: {taskHeight}px; display: flex; justify-content: flex-start; align-items: center; gap: 8px; padding: 0 8px; cursor: pointer; color: white;"
												>
													<div
														class="resize-handle-left"
														onmousedown={(e) => handleResizeStart(e, calendar, member, 'left')}
													></div>

													<h4
														style="color:white; margin: 0; text-overflow: ellipsis; overflow: hidden; white-space: nowrap; flex: 1; font-size: {compact &&
														taskHasOverlap
															? '0.7em'
															: '1em'};"
													>
														{calendar.title}
													</h4>
													{#if lanes.laneCount === 1 || !taskHasOverlap}
														<div
															style="gap: 5px; display: flex; justify-content: flex-end; gap: -10px;"
														>
															{#each calendar.hackers as hacker}
																<span class="task-avatar-list">
																	<Avatar
																		email={hacker.email}
																		option={{
																			showName: false,
																			size: 'md',
																			emptyFields: false,
																			circle: true,
																			tooltipEnabled: false
																		}}
																	/>
																</span>
															{/each}
														</div>
													{/if}

													<div
														class="resize-handle-right"
														onmousedown={(e) => handleResizeStart(e, calendar, member, 'right')}
													></div>
												</div>
											{/if}
										{/each}
									{/if}
								</td>
							{/each}
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
		<div class="card-footer d-flex align-items-center">
			<p class="m-0 text-secondary">
				{#if visibleDays.length > 0}
					Showing <span>{visibleStats.count}</span> tasks with
					<span>{Math.ceil(visibleStats.hours)}</span> hours
				{/if}
				<!-- | Total: <span>{calendarEvents?.length || 0}</span> assessments with <span>{calendarEvents?.reduce(
					(total, event) => total + (event.durationHours || 0),
					0
				).toFixed(1)}</span> hours -->
			</p>
		</div>
	</div>
{/if}

{#if showModal && selectedTask}
	<TaskModal
		task={selectedTask}
		{clickPosition}
		onClose={closeModal}
		onOpenFullView={openFullView}
		on:updateHackers={(e) => updateHackers(selectedTask, e)}
		on:colorchange={(e) => updateColor(selectedTask, e)}
		on:titlechange={(e) => updateTitle(selectedTask, e)}
		on:projectchange={(e) => updateProjects(selectedTask, e)}
		on:statuschange={(e) => updateStatus(selectedTask, e)}
	/>
{/if}

<style>
	.cell {
		padding: 0;
		position: relative;
	}

	/* Show avatars by default */
	.task {
		container-type: inline-size;
	}

	@container (max-width: 150px) {
		.task-avatar-list {
			display: none !important;
		}
	}

	/* Only show first avatar when space is limited */
	@container (min-width: 151px) and (max-width: 200px) {
		.task-avatar-list:nth-child(n + 1) {
			display: none !important;
		}
	}

	/* Show two avatars in medium-sized tasks */
	@container (min-width: 201px) and (max-width: 250px) {
		.task-avatar-list:nth-child(n + 3) {
			display: none !important;
		}
	}

	.table-responsive {
		height: 100%;
		max-height: 80vh;
	}
	.task {
		position: absolute;
		border-radius: 7px;
		border: 2px solid var(--tblr-border-color);
		left: 0;
		z-index: 10;
		user-select: none;
		-webkit-user-select: none;
		-moz-user-select: none;
		-ms-user-select: none;
		transition: all 0.2s ease;
		cursor: move;
	}

	.task-compact {
		border-width: 1px;
		border-radius: 4px;
	}

	th {
		text-align: center;
		vertical-align: middle;
	}

	.weekend {
		background-color: rgba(
			var(--tblr-muted-rgb),
			0.1
		); /* Mimicking tabler.io's bg-muted-lt class */
		min-width: 20px;
		width: 20px;
	}

	/* Hide day labels in weekend header cells to make them narrow */
	th.weekend {
		font-size: 0px;
	}

	.table {
		overflow-x: hidden;
	}

	.sticky-col {
		position: -webkit-sticky; /* For Safari */
		position: sticky;
		background-color: var(
			--tblr-body-bg
		); /* Background color is necessary to avoid content overlap */
		left: 0;
		z-index: 100; /* Ensure the sticky column is above other elements */
	}

	/* Add this if you want a border separation */
	.first-col {
		border-right: solid 1px var(--tblr-body-bg); /* Bootstrap's default border color */
	}

	.task-unselected {
		filter: brightness(0.7);
	}

	.task-selected {
		filter: brightness(1.2);
		z-index: 20; /* Raise selected task above others */
	}

	.task-dimmed {
		filter: brightness(0.7) opacity(0.8);
		z-index: 5; /* Lower than normal tasks */
	}

	/* Dim tasks during drag operations */
	.task-dimmed-drag {
		filter: brightness(0.6) opacity(0.5);
		z-index: 5; /* Lower than normal tasks */
		transition: filter 0.15s ease;
	}

	.task-done {
		opacity: 0.5;
		z-index: 1;
		/* border:none; */
	}

	/* Selection styling */
	.cell-selecting {
		background-color: rgba(59, 130, 246, 0.3) !important; /* Light blue selection color */
		position: relative;
		transition: background-color 0.1s ease;
	}

	/* Make sure weekend cells still show their color but with selection overlay */
	.cell-selecting.weekend {
		background-color: rgba(59, 130, 246, 0.3) !important;
	}

	/* Change cursor during selection */
	td {
		cursor: pointer;
	}

	/* When selection is active, show a horizontal resize cursor */
	.cell-selecting,
	tr:has(:global(.cell-selecting)) td:not(.first-col) {
		cursor: ew-resize !important;
	}

	/* Prevent text selection during drag */
	table {
		user-select: none;
	}

	tr td,
	tr {
		user-select: none;
		-webkit-user-select: none;
	}

	.avatar-container {
		display: inline-block;
		position: relative;
		margin-right: 5px;
		margin-bottom: 5px;
	}

	.overlay {
		position: absolute;
		top: 0;
		left: 0;
		width: 2em;
		height: 2em;
		background-color: rgba(0, 0, 0, 0.5);
		display: flex;
		justify-content: center;
		align-items: center;
		color: white;
		font-size: 16px;
		cursor: pointer;
		border-radius: 50%;
		z-index: 10;
	}

	.users-list-container {
		max-height: 400px;
		overflow-y: auto;
	}

	.cursor-pointer {
		cursor: pointer;
	}

	/* Task preview styling - improved */
	.task-preview {
		pointer-events: none;
		box-shadow: 0 6px 20px rgba(0, 0, 0, 0.35);
		opacity: 1 !important; /* Full opacity - not transparent */
		transition: none;
		position: fixed !important; /* Ensure fixed positioning */
		z-index: 2000; /* Very high z-index to appear above everything */
		transform: scale(1.02); /* Slightly larger to emphasize it's being dragged */
	}

	/* Body states during drag operations */
	body.dragging-active,
	body.resizing-active {
		cursor: move;
	}

	body.resizing-active {
		cursor: ew-resize;
	}

	/* Resize handles */
	.resize-handle-left,
	.resize-handle-right {
		position: absolute;
		top: 0;
		bottom: 0;
		width: 10px;
		cursor: ew-resize;
		z-index: 20;
	}

	.resize-handle-left {
		left: 0;
	}

	.resize-handle-right {
		right: 0;
	}

	/* Force cursor when starting resize operation */
	.resize-handle-left:active,
	.resize-handle-right:active {
		cursor: ew-resize !important;
	}
</style>
