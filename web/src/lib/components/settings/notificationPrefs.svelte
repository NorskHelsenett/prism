<script>
  // Per-user opt-out matrix for the two channels (in-app dropdown, web-push)
  // crossed with the two event kinds the dispatcher emits (new vulnerability,
  // new comment). The server treats unset === true, so a never-touched user
  // gets everything. Toggling here writes the explicit boolean.
  import { onMount } from 'svelte';
  import { Fetch } from '$lib/fetchUtil.js';

  /** @type {{inAppNewVuln: boolean, inAppNewComment: boolean, pushNewVuln: boolean, pushNewComment: boolean}} */
  let prefs = $state({
    inAppNewVuln: true,
    inAppNewComment: true,
    pushNewVuln: true,
    pushNewComment: true,
  });
  /** @type {string[]} */
  let swimlaneUsers = $state([]);
  let loaded = $state(false);
  let saving = $state(false);
  /** @type {Date | null} */
  let lastSavedAt = $state(null);

  onMount(async () => {
    const settings = await Fetch('/api/profile/preferences');
    if (settings) {
      swimlaneUsers = settings.swimlaneUsers ?? [];
      const np = settings.notificationPrefs ?? {};
      // Server omits unset keys so they decode as undefined; treat undefined
      // as "on" (the dispatcher's default).
      prefs = {
        inAppNewVuln: np.inAppNewVuln ?? true,
        inAppNewComment: np.inAppNewComment ?? true,
        pushNewVuln: np.pushNewVuln ?? true,
        pushNewComment: np.pushNewComment ?? true,
      };
    }
    loaded = true;
  });

  async function save() {
    saving = true;
    try {
      await Fetch('/api/profile/preferences', {
        method: 'PATCH',
        body: JSON.stringify({
          swimlaneUsers,
          notificationPrefs: prefs,
        }),
      });
      lastSavedAt = new Date();
    } finally {
      saving = false;
    }
  }

  /** @param {keyof typeof prefs} key */
  function toggle(key) {
    prefs[key] = !prefs[key];
    save();
  }
</script>

{#if loaded}
<div class="card">
  <div class="card-body">
    <h3 class="card-title">Notification preferences</h3>
    <p class="card-subtitle">Choose which events reach you, and on which channel. Everything is on by default. Web-push only applies to devices where you've enabled notifications.</p>

    <div class="table-responsive">
      <table class="table table-vcenter">
        <thead>
          <tr>
            <th></th>
            <th class="text-center">In-app</th>
            <th class="text-center">Web push</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>New vulnerability</td>
            <td class="text-center">
              <label class="form-check form-check-single form-switch m-0">
                <input class="form-check-input" type="checkbox" checked={prefs.inAppNewVuln} onchange={() => toggle('inAppNewVuln')} />
              </label>
            </td>
            <td class="text-center">
              <label class="form-check form-check-single form-switch m-0">
                <input class="form-check-input" type="checkbox" checked={prefs.pushNewVuln} onchange={() => toggle('pushNewVuln')} />
              </label>
            </td>
          </tr>
          <tr>
            <td>New comment</td>
            <td class="text-center">
              <label class="form-check form-check-single form-switch m-0">
                <input class="form-check-input" type="checkbox" checked={prefs.inAppNewComment} onchange={() => toggle('inAppNewComment')} />
              </label>
            </td>
            <td class="text-center">
              <label class="form-check form-check-single form-switch m-0">
                <input class="form-check-input" type="checkbox" checked={prefs.pushNewComment} onchange={() => toggle('pushNewComment')} />
              </label>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="text-muted small mt-2">
      {#if saving}
        Saving…
      {:else if lastSavedAt}
        Saved {lastSavedAt.toLocaleTimeString()}
      {/if}
    </div>
  </div>
</div>
{/if}
