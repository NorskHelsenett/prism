<script>
	import { accessLevels } from "$lib/userStore";

	let hoveredItem = $state(null);
  /**
   * @typedef {Object} Props
   * @property {string} [activeItem]
   */

  /** @type {Props} */
  let { activeItem = $bindable('') } = $props();

  const steps = [
    {
      title: 'Reported',
      description: 'The vulnerability is initially reported and awaits triage.'
    },
    {
      title: 'Validated',
      description: 'Confirmed as a valid issue by the security team.'
    },
    {
      title: 'In Progress',
      description: 'The vulnerability is officially handed over to the responsible team or individual for remediation. Currently being worked on to fix.'
    },
    {
      title: 'Rejected',
      description: 'Determined as a false positive or non-issue.'
    },
    {
      title: 'Resolved',
      description: 'The issue has been fully addressed, including all necessary follow-up actions.'
    }
  ];
</script>

{#if $accessLevels["/vulnerability"]?.write}
<ul class="steps steps-vertical">
	{#each steps as step, index}
		<li class="step-item cursor-pointer"
    onclick={() => activeItem = step.title}
    class:active={activeItem === step.title}
    >
			<div
				class="hover-border-effect padding"
				class:bg-primary-lt={hoveredItem === index}
				class:card-active={hoveredItem === index}
				onmouseover={() => (hoveredItem = index)}
				onmouseout={() => (hoveredItem = null)}
			>
				<div class="h4 m-0">{step.title}</div>
				<div class="text-secondary">{step.description}</div>
			</div>
		</li>
	{/each}
</ul>
{:else}
<ul class="steps steps-vertical">
	{#each steps as step, index}
		<li class="step-item"
    class:active={activeItem === step.title}
    >
			<div
				class="padding"
				class:bg-primary-lt={hoveredItem === index}
				class:card-active={hoveredItem === index}
			>
				<div class="h4 m-0">{step.title}</div>
				<div class="text-secondary">{step.description}</div>
			</div>
		</li>
	{/each}
</ul>
{/if}

<style>
  .padding {
    padding: 5px;
    transition: padding 0.3s ease-in-out;
  }
	.hover-border-effect {
		border: 1px solid #0000;
		transition: border 0.3s ease-in-out; /* Smooth transition for the border */
	}
	.hover-border-effect:hover {
		border: 1px solid var(--tblr-primary);
		border-radius: 0.25rem;
	}
  :global(div) {
    -webkit-box-sizing: border-box;
    -moz-box-sizing: border-box;
    box-sizing: border-box;
  }
</style>
