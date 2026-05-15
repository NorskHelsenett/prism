<script>
    import { createEventDispatcher } from 'svelte';

    /**
     * @typedef {Object} Props
     * @property {string} [value]
     * @property {number} [debounceInterval] - default debounce interval
     * @property {string} [placeholder]
     * @property {string} [id]
     * @property {boolean} [persisting]
     */

    /** @type {Props} */
    let {
        value = '',
        debounceInterval = 500,
        placeholder = '',
        id = '',
        persisting = false
    } = $props();

    const dispatch = createEventDispatcher();

    let timeout;

    // Debounce function
    function debounce(func, wait) {
        return function(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }

    // Emit the change after the specified delay
    const emitChange = debounce((newValue) => {
        dispatch('change', newValue);
    }, debounceInterval);

    function handleInput(event) {
        emitChange(event.target.value);
    }
</script>

<div class="form-floating mb-3">
    <input
        type="text"
        class="form-control"
        id={id}
        value={value}
        oninput={handleInput}>
    <label for={id}>{placeholder}</label>
    {#if persisting}
        <span class="input-icon-addon">
            <div class="spinner-border spinner-border-sm text-secondary" role="status"></div>
        </span>
    {/if}
</div>

