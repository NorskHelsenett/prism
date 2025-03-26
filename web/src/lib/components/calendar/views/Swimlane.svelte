<script>
	import Avatar from '$lib/components/Avatar.svelte';
	import TaskModal from './TaskModal.svelte';
	import { Fetch } from '$lib/fetchUtil';
	import { onMount } from 'svelte';

	let teams = [];
	let allUsers = [];
	let allTeams = [];
	let editMode = false;
	const { startDate, endDate } = calculateWeek();
	let loading = true;

	// Change the loading condition to only check if teams are loaded
	$: loading = !teams?.length;

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
	let originalTaskPosition = { dateFrom: null, dateTo: null, member: null };
	let cellWidth = 50; // Default cell width (will be measured on drag start)
	let weekendCellWidth = 20; // Width of weekend cells
	let originalTaskElement = null; // Add a reference to track the original element being dragged

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
		// Add scroll event listener
		window.addEventListener('scroll', handleScroll, true);

		return () => {
			document.removeEventListener('mouseup', handleMouseUp);
			document.removeEventListener('mousemove', handleDragMove);
			window.removeEventListener('scroll', handleScroll, true);
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
	let userFilter = '';
	$: filteredAvailableUsers = getAvailableUsers().filter(
		(user) =>
			user.email.toLowerCase().includes(userFilter.toLowerCase()) ||
			(user.name && user.name.toLowerCase().includes(userFilter.toLowerCase()))
	);

	$: filteredAvailableTeams = getAvailableTeams().filter(
		(team) =>
			team.name.toLowerCase().includes(userFilter.toLowerCase()) ||
			team.members.some((email) => email.toLowerCase().includes(userFilter.toLowerCase()))
	);

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

	let calendarEvents = [];
	export let reload = false;
	let days = generateDays(startDate, endDate);

	async function fetchCalendarEvents() {
		try {
			const events = await Fetch(
				`/api/planning?startDate=${formatDateToISO(startDate)}&endDate=${formatDateToISO(endDate)}`
			);

			// Process all events to add calculated properties
			calendarEvents = events.map((event) => {
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

	$: if (!reload) {
		fetchCalendarEvents();
	}

	function calculateTaskWidth(startDate, endDate) {
		// Convert string dates to Date objects if needed
		const start = new Date(startDate);
		const end = new Date(endDate);

		let totalWidth = 0;
		const regularDayWidth = 49; // Adjust based on your regular day width
		const weekendDayWidth = 19; // From your CSS for weekend cells

		for (let day = new Date(start); day <= end; day.setDate(day.getDate() + 1)) {
			const isWeekend = day.getDay() === 0 || day.getDay() === 6;
			totalWidth += isWeekend ? weekendDayWidth : regularDayWidth;
		}

		return totalWidth + 'px';
	}

	// Modal properties
	let showModal = false;
	let selectedTask = null;
	let selectedTaskElement = null;
	let lastClickedTaskId = null;
	let clickPosition = { x: 0, y: 0 };

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

		// Store reference to original element and hide it
		originalTaskElement = event.currentTarget;
		originalTaskElement.style.visibility = 'hidden';

		// Create and position the preview element
		createDragPreview(originalTaskElement, task);

		// Add global class to indicate dragging state
		document.body.classList.add('dragging-active');
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
		dragPreviewElement.style.opacity = '0.8'; // Slightly more opaque
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

    if (isResizing) {
        handleResizeMove(event);
        return;
    }

    if (!dragPreviewElement || !draggedTask) return;

    // Calculate movement in pixels
    const deltaX = event.clientX - dragStartX;
    const deltaY = event.clientY - dragStartY;

    // Find the cell under the cursor for snapping
    // Temporarily make the preview transparent for hit testing only
    dragPreviewElement.style.pointerEvents = 'none';
    const originalOpacity = dragPreviewElement.style.opacity;
    dragPreviewElement.style.opacity = '0';

    const elemBelow = document.elementFromPoint(event.clientX, event.clientY);

    // Restore opacity immediately
    dragPreviewElement.style.opacity = originalOpacity;

    const cell = elemBelow?.closest('td:not(.first-col)');

    if (cell) {
        const cellDate = cell.getAttribute('data-date');
        const cellMember = cell.getAttribute('data-member');

        if (cellDate && cellMember) {
            // Store the current target member for use when the drag ends
            dragTargetMember = cellMember;

            // Calculate new dates
            const daysDiff = calculateDaysDifference(dragStartDate, cellDate);

            // CHANGE: Always update the dates, even if daysDiff is 0
            const newStartDate = new Date(originalTaskPosition.dateFrom);
            newStartDate.setDate(newStartDate.getDate() + daysDiff);

            const newEndDate = new Date(originalTaskPosition.dateTo);
            newEndDate.setDate(newEndDate.getDate() + daysDiff);

            // Update the preview task
            draggedTask.dateFrom = newStartDate.toISOString().split('T')[0];
            draggedTask.dateTo = newEndDate.toISOString().split('T')[0];

            // Always position the preview at the appropriate cell
            positionDragPreviewAtCell(draggedTask.dateFrom, cellMember);
            dragPreviewElement.style.visibility = 'visible';
        }
    } else {
        // If not over a cell, just move the preview with the cursor
        const rect = dragPreviewElement.getBoundingClientRect();
        dragPreviewElement.style.top = rect.top + deltaY + 'px';
        dragPreviewElement.style.left = rect.left + deltaX + 'px';
        dragPreviewElement.style.visibility = 'visible';

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

			// With fixed positioning, we use viewport coordinates directly
			dragPreviewElement.style.top = cellRect.top + 'px';
			dragPreviewElement.style.left = cellRect.left + 'px';
			dragPreviewElement.style.width = taskDuration;
			dragPreviewElement.style.visibility = 'visible';
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

		// Store and hide the original element
		originalTaskElement = event.target.closest('.task');
		originalTaskElement.style.visibility = 'hidden';

		// Create and position the preview
		createDragPreview(originalTaskElement, task);

		document.body.classList.add('resizing-active');
	}

	// Handle resizing movement - Fixed visibility issues
	function handleResizeMove(event) {
		if (!isResizing || !dragPreviewElement) return;

		// Hide preview temporarily to detect cells underneath
		// Use opacity instead of visibility for hit testing
		const originalOpacity = dragPreviewElement.style.opacity;
		dragPreviewElement.style.opacity = '0';

		const elemBelow = document.elementFromPoint(event.clientX, event.clientY);

		// Restore opacity immediately
		dragPreviewElement.style.opacity = originalOpacity;

		const cell = elemBelow?.closest('td:not(.first-col)');

		if (cell) {
			const cellDate = cell.getAttribute('data-date');
			const cellMember = cell.getAttribute('data-member');

			if (cellDate && cellMember === originalTaskPosition.member) {
				// Calculate new dates based on resize direction
				if (resizeDirection === 'left') {
					// Update start date
					draggedTask.dateFrom = cellDate;
				} else if (resizeDirection === 'right') {
					// Update end date
					draggedTask.dateTo = cellDate;
				}

				// Ensure end date isn't before start date
				const startDate = new Date(draggedTask.dateFrom);
				const endDate = new Date(draggedTask.dateTo);

				if (endDate < startDate) {
					if (resizeDirection === 'left') {
						draggedTask.dateFrom = draggedTask.dateTo;
					} else {
						draggedTask.dateTo = draggedTask.dateFrom;
					}
				}

				// Update width and position of preview
				const newWidth = calculateTaskWidth(draggedTask.dateFrom, draggedTask.dateTo);
				dragPreviewElement.style.width = newWidth;

				if (resizeDirection === 'left') {
					// When resizing from left, update the position too
					positionDragPreviewAtCell(draggedTask.dateFrom, originalTaskPosition.member);
				}
			}
		}

		// Always ensure the preview is visible after positioning
		dragPreviewElement.style.visibility = 'visible';

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
            // Remove the original member
            updatedHackers = updatedHackers.filter((h) => h.email !== dragStartMember);
            
            // Check if target member is already in the list to avoid duplicates
            if (!updatedHackers.some(h => h.email === dragTargetMember)) {
              updatedHackers.push({ email: dragTargetMember });
            }
          }

          // Additional uniqueness check - create a clean list with no duplicates
          const uniqueEmails = new Set();
          updatedHackers = updatedHackers.filter(h => {
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
      // Restore visibility of the original element
      if (originalTaskElement) {
        originalTaskElement.style.visibility = '';
        originalTaskElement = null;
      }

      // Clean up
      if (dragPreviewElement) {
        dragPreviewElement.remove();
        dragPreviewElement = null;
      }

      isDragging = false;
      isResizing = false;
      draggedTask = null;
      dragTargetMember = null; // Reset the target member
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
	function getEventsForDayAndMember(dayDate, memberEmail) {
		if (!calendarEvents?.length) return null;

		if (formatDateToISO(startDate) === dayDate) {
			const matchingEvent = calendarEvents.find((event) => {
				const eventStartDate = new Date(event.dateFrom).toISOString().split('T')[0];
				// Only show the task on its start date and if member is assigned
				return eventStartDate <= dayDate && event.hackers.some((h) => h.email === memberEmail);
			});
			return matchingEvent || null;
		}

		// Find the first event that starts on this exact day for this member
		// This prevents the same task from appearing on multiple days
		const matchingEvent = calendarEvents.find((event) => {
			const eventStartDate = new Date(event.dateFrom).toISOString().split('T')[0];

			// Only show the task on its start date and if member is assigned
			return eventStartDate === dayDate && event.hackers.some((h) => h.email === memberEmail);
		});

		return matchingEvent || null;
	}

	// Add this function to handle scrolling during drag operations
	function handleScroll() {
		if (!isDragging && !isResizing) return;
		if (!dragPreviewElement || !draggedTask) return;

		// Update preview position based on the current target
		if (isDragging && dragTargetMember) {
			positionDragPreviewAtCell(draggedTask.dateFrom, dragTargetMember);
		} else if (isResizing) {
			// For resizing, always use the original member
			positionDragPreviewAtCell(draggedTask.dateFrom, originalTaskPosition.member);
		}
	}
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
				on:click={toggleEditMode}
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
									<i class="overlay ti ti-x rounded-circle" on:click={() => removeUser(member)}></i>
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
												on:click={() => addTeam(getAddableTeamMembers(team))}
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
											on:click={() => addUser(user.email)}
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
			<table class="table table-vcenter card-table">
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
									class:weekend={day.isWeekend}
									on:mousedown={(e) => handleMouseDown(e, day, member)}
									on:mousemove={(e) => handleMouseMove(e, day, member)}
									on:dragstart={(e) => e.preventDefault()}
									data-date={day.date}
									data-member={member}
								>
									<!-- Call the function once and store the result -->
									{#if calendarEvents?.length}
										{@const calendar = calendarEvents?.length
											? getEventsForDayAndMember(day.date, member)
											: null}
										{#if calendar}
											<span class:weekend={day.isWeekend} style="position: relative;">
												<!-- svelte-ignore a11y-click-events-have-key-events -->
												<!-- svelte-ignore a11y-no-static-element-interactions -->
												<div
													class="task planningid-{calendar.id}"
													on:click={(e) => handleTaskClick(e, calendar)}
													on:mousedown={(e) => handleTaskDragStart(e, calendar, member)}
													style="width: {calendar.width}; background-color: {calendar.color}; display: flex; justify-content: flex-start; align-items: center; gap: 8px; padding: 0 8px; cursor: pointer; color: white;"
												>
													<div
														class="resize-handle-left"
														on:mousedown={(e) => handleResizeStart(e, calendar, member, 'left')}
													></div>

													<h4 style="color:white; margin: 0; text-overflow: ellipsis; overflow: hidden; white-space: nowrap; flex: 1;">
														{calendar.title}
													</h4>
                          <!-- <svg  xmlns="http://www.w3.org/2000/svg"  width="24"  height="24"  viewBox="0 0 24 24"  fill="none"  stroke="currentColor"  stroke-width="2"  stroke-linecap="round"  stroke-linejoin="round"  class="icon icon-tabler icons-tabler-outline icon-tabler-letter-i"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M12 4l0 16" /></svg> -->
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

													<div
														class="resize-handle-right"
														on:mousedown={(e) => handleResizeStart(e, calendar, member, 'right')}
													></div>
												</div>
											</span>
										{/if}
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
				Showing <span>{calendarEvents?.length || 0}</span> assessments with a total of {calendarEvents?.reduce(
					(total, event) => total + (event.durationHours || 0),
					0
				)} hours
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
	/>
{/if}

<style>

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
    .task-avatar-list:nth-child(n+2) {
      display: none !important;
    }
  }

  /* Show two avatars in medium-sized tasks */
  @container (min-width: 201px) and (max-width: 250px) {
    .task-avatar-list:nth-child(n+3) {
      display: none !important;
    }
  }

	.table-responsive {
		height: 100%;
		max-height: 80vh;
	}
	.task {
		height: 55px;
		position: absolute;
		top: 5px;
		border-radius: 7px;
		border: 3px solid #1b2432;
		left: 0;
		z-index: 10;
		user-select: none;
		-webkit-user-select: none;
		-moz-user-select: none;
		-ms-user-select: none;
		/*transition: all 0.2s ease;  Changed to 'all' to animate selection highlighting */
		cursor: move; /* Show move cursor for draggable tasks */
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
		font-size: 0px;
		min-width: 20px;
		width: 20px;
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
	tr:has(.cell-selecting) td:not(.first-col) {
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
		box-shadow: 0 4px 15px rgba(0, 0, 0, 0.5);
		opacity: 0.8;
		transition: none;
		position: fixed !important; /* Ensure fixed positioning */
		z-index: 1000;
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
