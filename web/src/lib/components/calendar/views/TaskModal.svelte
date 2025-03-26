<script>
  import Icon from '$lib/components/Icon.svelte';
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

  function updateHackers(event) {
    if (!event) return;
    task.hackers = event.detail
    console.log('task.hackers', task.hackers);
    dispatch('updateHackers', task.hackers);
  }
</script>

<svelte:window on:click={closeColorPicker} />

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
  
  <div class="modal-content">
    <div class="icon-column">
      <svg  xmlns="http://www.w3.org/2000/svg"  width="24"  height="24"  viewBox="0 0 24 24"  fill="none"  stroke="currentColor"  stroke-width="2"  stroke-linecap="round"  stroke-linejoin="round"  class="icon icon-tabler icons-tabler-outline icon-tabler-clock-hour-4"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M12 12m-9 0a9 9 0 1 0 18 0a9 9 0 1 0 -18 0" /><path d="M12 12l3 2" /><path d="M12 7v5" /></svg>
    </div>
    <div class="text-column">
      <p><strong>Date:</strong> {task.dateFrom} to {task.dateTo}</p>
      <p><strong>Work Order:</strong> {task.workorder || 'N/A'}</p>
      <p><strong>Status:</strong> {task.status || 'N/A'}</p>
    </div>
    <div class="icon-column">
      <svg  xmlns="http://www.w3.org/2000/svg"  width="24"  height="24"  viewBox="0 0 24 24"  fill="none"  stroke="currentColor"  stroke-width="2"  stroke-linecap="round"  stroke-linejoin="round"  class="icon icon-tabler icons-tabler-outline icon-tabler-mailbox"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M10 21v-6.5a3.5 3.5 0 0 0 -7 0v6.5h18v-6a4 4 0 0 0 -4 -4h-10.5" /><path d="M12 11v-8h4l2 2l-2 2h-4" /><path d="M6 15h1" /></svg>
    </div>
    <div class="text-column">
      <p><strong>Requester:</strong> {task.requester || 'N/A'}</p>
      <p><strong>Responsible:</strong> {task.responsible_hacker || 'N/A'}</p>
      <p><strong>Estimate:</strong> {task.estimate || 0} hours</p>
      <p><strong>Project:</strong> {task.projects?.map(p => p.name).join(', ') || 'N/A'}</p>
    </div>
    {#if task.note}
    <div class="icon-column">
      <svg  xmlns="http://www.w3.org/2000/svg"  width="24"  height="24"  viewBox="0 0 24 24"  fill="none"  stroke="currentColor"  stroke-width="2"  stroke-linecap="round"  stroke-linejoin="round"  class="icon icon-tabler icons-tabler-outline icon-tabler-notes"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M5 3m0 2a2 2 0 0 1 2 -2h10a2 2 0 0 1 2 2v14a2 2 0 0 1 -2 2h-10a2 2 0 0 1 -2 -2z" /><path d="M9 7l6 0" /><path d="M9 11l6 0" /><path d="M9 15l4 0" /></svg>
    </div>
    <div class="text-column">
      <p><strong>Note:</strong> {task.note}</p>
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
    margin-bottom: 16px;
    border-bottom: 1px solid var(--tblr-border-color, #e6e7e9);
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
    grid-template-columns: auto 1fr;
    gap: 12px;
  }

  .modal-content p {
    margin:0;
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
</style>