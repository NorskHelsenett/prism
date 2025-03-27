<script>
  import Icon from '$lib/components/Icon.svelte';
	import Markdown from '../../Markdown.svelte';
	import Avatarlist from '../Avatarlist.svelte';
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  const dispatch = createEventDispatcher();

  // Props
  export let task; // The task to display
  export let clickPosition = { x: 0, y: 0 }; // Only need click coordinates, not final position
  export let onClose; // Function to close the modal
  export let onOpenFullView; // Function to open task in full view
  
  // Editable title state
  let editableTitle = task.title || '';
  let titleDebounceTimer;
  let titleInputRef;
  let isEditingTitle = false;
  
  let modalRef;
  let modalPosition = { top: '0px', left: '0px' };
  let showColorPicker = false;
  let selectedColor = task.color || '#206bc4'; // Default color or from task
  
  const colors = [
    '#206bc4', // blue
    '#4299e1', // light blue
    '#2fb344', // green
    '#f76707', // orange
    '#d63939', // red
    '#ae3ec9', // purple
    '#fbbf24', // yellow
    '#64748b', // slate
    '#0ca678', // teal
    '#6366f1'  // indigo
  ];
  
  function selectColor(color) {
    selectedColor = color;
    showColorPicker = false;
    // Here you would typically save the color to the task
    task.color = color;
    dispatch('colorchange', { color, task }); // Notify parent
  }
  
  // Calculate position on mount and when clickPosition changes
  $: if (clickPosition && modalRef) {
    calculatePosition();
  }
  
  function calculatePosition() {
    const mousePosition = { x: clickPosition.x, y: clickPosition.y };
    
    // Use requestAnimationFrame instead of setTimeout for better performance
    requestAnimationFrame(() => {
      if (!modalRef) return;
      
      const modalRect = modalRef.getBoundingClientRect();
      const viewportWidth = window.innerWidth;
      const viewportHeight = window.innerHeight;
      
      // Calculate available space in different directions
      const spaceAbove = clickPosition.y;
      const spaceBelow = viewportHeight - clickPosition.y;
      const spaceLeft = clickPosition.x;
      const spaceRight = viewportWidth - clickPosition.x;
      
      let top, left;
      
      // Positioning logic (same as before)
      if (spaceRight >= modalRect.width + 20) {
        left = clickPosition.x + window.scrollX + 15;
        top = clickPosition.y + window.scrollY - (modalRect.height / 2);
      } else if (spaceLeft >= modalRect.width + 20) {
        left = clickPosition.x + window.scrollX - modalRect.width - 15;
        top = clickPosition.y + window.scrollY - (modalRect.height / 2);
      } else if (spaceBelow >= modalRect.height + 15) {
        top = clickPosition.y + window.scrollY + 15;
        left = clickPosition.x + window.scrollX - (modalRect.width / 2);
      } else if (spaceAbove >= modalRect.height + 15) {
        top = clickPosition.y + window.scrollY - modalRect.height - 15;
        left = clickPosition.x + window.scrollX - (modalRect.width / 2);
      } else {
        // Default center position
        top = (viewportHeight - modalRect.height) / 2 + window.scrollY;
        left = (viewportWidth - modalRect.width) / 2 + window.scrollX;
      }
      
      // Ensure the modal stays within viewport boundaries
      left = Math.max(10 + window.scrollX, Math.min(left, viewportWidth - modalRect.width - 10 + window.scrollX));
      top = Math.max(10 + window.scrollY, Math.min(top, viewportHeight - modalRect.height - 10 + window.scrollY));
      
      // Update modal position
      modalPosition = { top: top + 'px', left: left + 'px' };
    });
  }

  // Click outside action
  function clickOutside(node) {
    const handleClick = (event) => {
      if (!node.contains(event.target)) {
        node.dispatchEvent(new CustomEvent('outclick'));
      }
    };
    
    document.addEventListener('click', handleClick, true);
    
    return {
      destroy() {
        document.removeEventListener('click', handleClick, true);
      }
    };
  }
  
  // Close color picker when clicking outside
  function closeColorPicker(event) {
    if (showColorPicker && !event.target.closest('.color-selector') && 
        !event.target.closest('.color-overlay')) {
      showColorPicker = false;
    }
  }

  // Debounce function for title updates
  function updateTitle() {
    clearTimeout(titleDebounceTimer);
    titleDebounceTimer = setTimeout(() => {
      saveTitle();
    }, 500); // 500ms debounce delay
  }
  
  // Save the title to the task and dispatch event
  function saveTitle() {
    if (task.title !== editableTitle) {
      task.title = editableTitle;
      dispatch('titlechange', { title: editableTitle, task });
    }
  }
  
  // Handle title focus out
  function handleTitleFocusOut() {
    isEditingTitle = false;
    clearTimeout(titleDebounceTimer);
    saveTitle();
  }
  
  // Handle window beforeunload to save changes
  function handleBeforeUnload(event) {
    if (titleDebounceTimer) {
      clearTimeout(titleDebounceTimer);
      saveTitle();
    }
  }
  
  onMount(() => {
    window.addEventListener('beforeunload', handleBeforeUnload);
  });
  
  onDestroy(() => {
    // Clean up timers and event listeners
    if (titleDebounceTimer) {
      clearTimeout(titleDebounceTimer);
      saveTitle();
    }
    window.removeEventListener('beforeunload', handleBeforeUnload);
  });

  // Project selection state
  let projects = [];
  let showProjectDropdown = false;
  let isLoadingProjects = false;
  let projectError = null;
  let selectedProject = task.projects?.length > 0 ? task.projects[0] : null;

  // Fetch projects from API
  async function fetchProjects() {
    try {
      isLoadingProjects = true;
      projectError = null;
      const response = await fetch('/api/project/all');
        
      if (!response.ok) {
        throw new Error(`Failed to fetch projects: ${response.status}`);
      }
      
      projects = await response.json();
    } catch (error) {
      console.error('Error fetching projects:', error);
      projectError = error.message;
    } finally {
      isLoadingProjects = false;
    }
  }

  // Handle project selection
  function selectProject(project) {
    selectedProject = { id: project.ID, name: project.ProjectName };
    // Replace current projects with the selected one
    task.projects = [selectedProject];
    showProjectDropdown = false;
    
    // Notify parent component
    dispatch('projectchange', { projects: task.projects, task });
  }

  // Check if project is currently selected
  function isProjectSelected(projectId) {
    return selectedProject && selectedProject.id === projectId;
  }

  // Toggle project dropdown and load projects if needed
  function toggleProjectDropdown(event) {
    event.stopPropagation();
    showProjectDropdown = !showProjectDropdown;
    
    if (showProjectDropdown && projects.length === 0) {
      fetchProjects();
    }
  }

  // Close project dropdown when clicking outside
  function closeProjectDropdown(event) {
    if (showProjectDropdown && 
        !event.target.closest('.project-select-container')) {
      showProjectDropdown = false;
    }
  }
  
  onMount(() => {
    window.addEventListener('beforeunload', handleBeforeUnload);
    // Initialize projects if task already has projects
    if (task.projects && task.projects.length > 0) {
      // We might want to fetch all projects to ensure we have the full data
      fetchProjects();
    }
  });

  // Calculate workdays between two dates (excluding weekends)
  function getWorkdayCount(startDate, endDate) {
    // Convert string dates to Date objects
    const start = new Date(startDate);
    const end = new Date(endDate);
    
    // Initialize counter
    let count = 0;
    
    // Clone start date to avoid modifying it
    const currentDate = new Date(start);
    
    // Loop through all days
    while (currentDate <= end) {
      // Check if current day is not a weekend (0 = Sunday, 6 = Saturday)
      const dayOfWeek = currentDate.getDay();
      if (dayOfWeek !== 0 && dayOfWeek !== 6) {
        count++;
      }
      
      // Move to next day
      currentDate.setDate(currentDate.getDate() + 1);
    }
    
    return count;
  }

  // Calculate estimate based on hackers and workdays
  function calculateEstimate(hackers, dateFrom, dateTo) {
    const hackerCount = hackers?.length || 0;
    
    // If no hackers or no date range, return 0
    if (hackerCount === 0 || !dateFrom || !dateTo) {
      return 0;
    }
    
    const workdayCount = getWorkdayCount(dateFrom, dateTo);
    const hoursPerDay = 7.5;
    
    return Math.floor(workdayCount * hoursPerDay * hackerCount);
  }

  // Reactive statement to update estimate when relevant data changes
  $: if (task && task.dateFrom && task.dateTo && task.hackers) {
    task.estimate = calculateEstimate(task.hackers, task.dateFrom, task.dateTo);
  }

  // Define Kanban statuses
  const kanbanStatuses = [
    { id: 'todo', label: 'To Do', color: 'bg-secondary' },
    { id: 'inprogress', label: 'In Progress', color: 'bg-primary' },
    { id: 'done', label: 'Done', color: 'bg-success' },
    { id: 'blocked', label: 'Blocked', color: 'bg-danger' }
  ];

  // Initialize status if not set
  task.status = task.status || 'todo';

  // Handle status change
  function changeStatus(newStatus) {
    if (task.status !== newStatus) {
      task.status = newStatus;
      dispatch('statuschange', { status: newStatus, task });
    }
  }

  function updateHackers(event) {
    if (!event) return;
    task.hackers = event.detail
    console.log('task.hackers', task.hackers);
    
    // Update estimate when hackers change
    task.estimate = calculateEstimate(task.hackers, task.dateFrom, task.dateTo);
    
    dispatch('updateHackers', task.hackers);
  }
</script>

<svelte:window on:click={(e) => { closeColorPicker(e); closeProjectDropdown(e); }} />

<div class="task-modal" 
     bind:this={modalRef}
     style="top: {modalPosition.top}; left: {modalPosition.left};"
     use:clickOutside on:outclick={onClose}>
  <button class="btn-expand" on:click={() => onOpenFullView(task.id)}>
    <Icon icon="arrows-maximize" />
  </button>
  
  <div class="modal-header">
    <div class="title-container">
      <div class="color-selector" 
           style="background-color: {selectedColor};" 
           on:click={() => showColorPicker = !showColorPicker}></div>
      
      {#if isEditingTitle}
        <input 
          type="text" 
          class="title-input"
          bind:this={titleInputRef}
          bind:value={editableTitle}
          on:input={updateTitle}
          on:blur={handleTitleFocusOut}
          on:keydown={(e) => e.key === 'Enter' && handleTitleFocusOut()}
          autofocus
        />
      {:else}
        <h3 on:click={() => { isEditingTitle = true; setTimeout(() => titleInputRef?.focus(), 0); }}>{task.title}</h3>
      {/if}
    </div>
    
  </div>
  
  {#if showColorPicker}
    <div class="color-overlay">
      <div class="color-options">
        {#each colors as color}
          <div class="color-option" 
               style="background-color: {color};" 
               on:click={() => selectColor(color)}></div>
        {/each}
      </div>
    </div>
  {/if}

  <div class="btn-group w-100" role="group">
    {#each kanbanStatuses as status}
      <input type="radio" 
             class="btn-check" 
             name="task-status" 
             id="status-{status.id}" 
             autocomplete="off"
             checked={task.status === status.id}
             on:change={() => changeStatus(status.id)}>
      <label for="status-{status.id}" 
             class="btn {status.color} {task.status !== status.id ? 'text-muted' : ''}">
        {status.label}
      </label>
    {/each}
  </div>
  
  <div class="modal-content">
    <div class="icon-column">
      <svg  xmlns="http://www.w3.org/2000/svg"  width="24"  height="24"  viewBox="0 0 24 24"  fill="none"  stroke="currentColor"  stroke-width="2"  stroke-linecap="round"  stroke-linejoin="round"  class="icon icon-tabler icons-tabler-outline icon-tabler-clock-hour-4"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M12 12m-9 0a9 9 0 1 0 18 0a9 9 0 1 0 -18 0" /><path d="M12 12l3 2" /><path d="M12 7v5" /></svg>
    </div>
    <div class="text-column">
      <p><strong>Date:</strong> {task.dateFrom} to {task.dateTo}</p>
      <p><strong>Work Order:</strong> {task.workorder || 'N/A'}</p>
    </div>
    <div class="icon-column">
      <svg  xmlns="http://www.w3.org/2000/svg"  width="24"  height="24"  viewBox="0 0 24 24"  fill="none"  stroke="currentColor"  stroke-width="2"  stroke-linecap="round"  stroke-linejoin="round"  class="icon icon-tabler icons-tabler-outline icon-tabler-mailbox"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M10 21v-6.5a3.5 3.5 0 0 0 -7 0v6.5h18v-6a4 4 0 0 0 -4 -4h-10.5" /><path d="M12 11v-8h4l2 2l-2 2h-4" /><path d="M6 15h1" /></svg>
    </div>
    <div class="text-column">
      <p><strong>Estimate:</strong> {task.estimate || 0} hours</p>
      
      <div class="mb-2 project-container">
        <strong>Project:</strong>
        <div class="project-select-container">
          <div class="form-select" on:click={toggleProjectDropdown}>
            {#if selectedProject}
              <div class="selected-project">{selectedProject.name}</div>
            {:else}
              <div class="placeholder">Select project</div>
            {/if}
          </div>
          
          {#if showProjectDropdown}
            <div class="project-dropdown">
              {#if isLoadingProjects}
                <div class="project-loading">Loading projects...</div>
              {:else if projectError}
                <div class="project-error">{projectError}</div>
              {:else if projects.length === 0}
                <div class="no-projects">No projects available</div>
              {:else}
                {#each projects as project}
                  <div 
                    class="project-option {isProjectSelected(project.ID) ? 'selected' : ''}"
                    on:click={() => selectProject(project)}
                  >
                    {project.ProjectName}
                  </div>
                {/each}
              {/if}
            </div>
          {/if}
        </div>
      </div>
    </div>
    {#if task.note}
    <div class="icon-column">
      <svg  xmlns="http://www.w3.org/2000/svg"  width="24"  height="24"  viewBox="0 0 24 24"  fill="none"  stroke="currentColor"  stroke-width="2"  stroke-linecap="round"  stroke-linejoin="round"  class="icon icon-tabler icons-tabler-outline icon-tabler-notes"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M5 3m0 2a2 2 0 0 1 2 -2h10a2 2 0 0 1 2 2v14a2 2 0 0 1 -2 2h-10a2 2 0 0 1 -2 -2z" /><path d="M9 7l6 0" /><path d="M9 11l6 0" /><path d="M9 15l4 0" /></svg>
    </div>
    <div class="text-column">
      <p style="max-height: 6em;"><strong>Note:</strong> <Markdown markdown={task.note} /></p>
    </div>
    {/if}
    <div class="icon-column">
      <svg  xmlns="http://www.w3.org/2000/svg"  width="24"  height="24"  viewBox="0 0 24 24"  fill="none"  stroke="currentColor"  stroke-width="2"  stroke-linecap="round"  stroke-linejoin="round"  class="icon icon-tabler icons-tabler-outline icon-tabler-users-group"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M10 13a2 2 0 1 0 4 0a2 2 0 0 0 -4 0" /><path d="M8 21v-1a2 2 0 0 1 2 -2h4a2 2 0 0 1 2 2v1" /><path d="M15 5a2 2 0 1 0 4 0a2 2 0 0 0 -4 0" /><path d="M17 10h2a2 2 0 0 1 2 2v1" /><path d="M5 5a2 2 0 1 0 4 0a2 2 0 0 0 -4 0" /><path d="M3 13v-1a2 2 0 0 1 2 -2h2" /></svg>
    </div>
    <div class="text-column">
      <strong>Hackers:</strong>
      <div class="avatars">
        <Avatarlist hackers={task.hackers} on:updateHackers="{e => updateHackers(e)}"/>
      </div>
    </div>
  </div>
</div>

<style>
  .task-modal {
    background: var(--tblr-body-bg, #fff);
    border: 1px solid var(--tblr-border-color, #e6e7e9);
    border-radius: 12px;
    box-shadow: 0 8px 30px rgba(0, 0, 0, 0.12);
    padding: 24px;
    width: 360px;
    z-index: 1000;
    max-width: 90vw;
    position: fixed;
    overflow: hidden;
  }
  
  .modal-header {
    /* margin-bottom: 16px; */
    /* border-bottom: 1px solid var(--tblr-border-color, #e6e7e9); */
    padding-bottom: 16px;
    padding-right: 32px;
    padding-left: 0;
  }
  
  .title-container {
    display: grid;
    grid-template-columns: auto 1fr;
    align-items: center;
    gap: 12px;
  }
  
  .modal-header h3 {
    margin: 0;
    font-size: 1.3rem;
    font-weight: 600;
    cursor: pointer;
    padding: 4px 8px;
    border-radius: 4px;
  }
  
  .modal-header h3:hover {
    background-color: rgba(0, 0, 0, 0.05);
  }
  
  .btn-expand {
    position: absolute;
    top: 16px;
    right: 16px;
    background: none;
    border: none;
    cursor: pointer;
    padding: 6px;
    width: 32px;
    height: 32px;
    color: var(--tblr-muted, #6c757d);
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 4px;
  }
  
  .btn-expand:hover {
    color: var(--tblr-primary, #206bc4);
    background-color: rgba(32, 107, 196, 0.1);
  }
  
  .modal-content {
    font-size: 0.95rem;
    display: grid;
    grid-template-columns: 36px 1fr;
    gap: 12px;
    width: 100%;
    overflow: hidden;
  }

  .modal-content p {
    margin: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  .icon-column {
    width: 36px;
    min-width: 36px;
    display: flex;
    align-items: flex-start;
    justify-content: center;
  }
  
  .text-column {
    overflow: hidden;
    min-width: 0; /* Important for overflow to work */
  }
  
  .avatars {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 8px;
  }
  
  /* Color selector styling */
  .color-selector {
    width: 20px;
    height: 20px;
    margin-left: 5px;
    border-radius: 50%;
    cursor: pointer;
    border: 2px solid white;
    box-shadow: 0 0 0 1px var(--tblr-border-color);
  }
  
  .color-overlay {
    position: absolute;
    top: 60px;
    left: 24px;
    background-color: white;
    border-radius: 8px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    padding: 12px;
    z-index: 1001;
    border: 1px solid var(--tblr-border-color);
  }
  
  .color-options {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    width: 152px;
  }
  
  .color-option {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    cursor: pointer;
    border: 2px solid white;
    box-shadow: 0 0 0 1px var(--tblr-border-color);
    transition: transform 0.2s ease;
  }
  
  .color-option:hover {
    transform: scale(1.2);
  }

  .title-input {
    font-size: 1.3rem;
    font-weight: 600;
    width: 100%;
    padding: 4px 8px;
    border-radius: 4px;
    border: 1px solid var(--tblr-border-color, #e6e7e9);
    background-color: var(--tblr-bg-surface, #fff);
    margin: 0;
  }
  
  .title-input:focus {
    outline: 2px solid var(--tblr-primary, #206bc4);
    border-color: var(--tblr-primary, #206bc4);
  }

  /* Project selector styles - fixed positioning */
  .project-container {
    position: relative;
    display: flex;
  }
  
  .project-select-container {
    position: relative;
    /* margin-top: 4px; */
    border:none; 
    width: 100%;
  }
  
  .form-select {
    padding: 0;
    padding-left: 5px;
    padding-top: 2px;
    display: flex;
    align-items: center;
    width: 100%;
    /* padding: 0.4375rem 0.75rem; */
    font-size: 0.875rem;
    font-weight: 400;
    line-height: 1.4285714;
    color: var(--tblr-body-color, #1e293b);
    background-color: var(--tblr-bg-forms, #fff);
    background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 16 16'%3e%3cpath fill='none' stroke='%23dadcde' stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M2 5l6 6 6-6'/%3e%3c/svg%3e");
    background-repeat: no-repeat;
    background-position: right 0.75rem center;
    background-size: 16px 12px;
    /* border: 1px solid var(--tblr-border-color, #dadcde); */
    border-radius: 4px;
    border:none;
    transition: border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out;
    appearance: none;
    cursor: pointer;
  }
  
  .form-select:hover {
    border-color: var(--tblr-primary, #206bc4);
  }
  
  .selected-project, .placeholder {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  
  .placeholder {
    color: var(--tblr-muted, #6c757d);
  }
  
  .project-dropdown {
    position: fixed;
    width: calc(100% - 32px - 36px);
    max-width: 280px;
    z-index: 1500;
    max-height: 200px;
    overflow-y: auto;
    margin-top: 2px;
    background-color: var(--tblr-bg-surface, #fff);
    border: 1px solid var(--tblr-border-color, #dadcde);
    border-radius: 4px;
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
  }
  
  .project-option {
    padding: 0.375rem 0.75rem;
    cursor: pointer;
    font-size: 0.875rem;
  }
  
  .project-option:hover {
    background-color: rgba(32, 107, 196, 0.1);
  }
  
  .project-option.selected {
    background-color: rgba(32, 107, 196, 0.1);
    color: var(--tblr-primary, #206bc4);
    font-weight: 500;
  }
  
  .project-loading,
  .project-error,
  .no-projects {
    padding: 0.75rem;
    text-align: center;
    color: var(--tblr-muted, #6c757d);
    font-size: 0.875rem;
  }
  
  .project-error {
    color: var(--tblr-danger, #d63939);
  }

  .btn-group {
    display: flex;
    border-radius: 4px;
    overflow: hidden;
    margin-bottom: 16px;
  }
  
  .btn-group .btn {
    flex: 1;
    font-size: 0.75rem;
    padding: 0.25rem 0.5rem;
    border-radius: 0;
    border: 1px solid var(--tblr-border-color, #e6e7e9);
    transition: all 0.2s ease;
  }
  
  .btn-group .btn:first-child {
    border-top-left-radius: 4px;
    border-bottom-left-radius: 4px;
  }
  
  .btn-group .btn:last-child {
    border-top-right-radius: 4px;
    border-bottom-right-radius: 4px;
  }
  
  .btn-check {
    position: absolute;
    clip: rect(0, 0, 0, 0);
    pointer-events: none;
  }
  
  .btn-check:checked + .btn {
    font-weight: 500;
    opacity: 1;
    color: #fff;
  }
  
  .btn-check:not(:checked) + .btn {
    background-color: var(--tblr-bg-surface, #fff) !important;
    color: var(--tblr-muted, #6c757d) !important;
  }
  
  .btn-check:not(:checked) + .btn:hover {
    background-color: rgba(0, 0, 0, 0.05) !important;
  }
  
  /* Remove previous styling that's no longer needed */
  .status-buttons {
    display: none;
  }
</style>