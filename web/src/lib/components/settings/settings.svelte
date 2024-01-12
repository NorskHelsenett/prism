<script>
	import Export from "./export.svelte";

</script>
<div class="card-body">
  <h3 class="card-title mt-4">Slack</h3>
  <p class="card-subtitle">Enable slack integration which will post a new slack message for each new vulnerability found.</p>
  <div>
    <label class="form-check form-switch form-switch-lg">
      <input class="form-check-input" type="checkbox">
      <span class="form-check-label form-check-label-on">Slack integration is now turned on</span>
      <span class="form-check-label form-check-label-off">Slack integration disabled</span>
    </label>
  </div>
</div>

<Export />