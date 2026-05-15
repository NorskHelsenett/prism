<script>
	import Export from "./export.svelte";
	import Slack from "./slack.svelte";
	import DatabaseCleanup from "./databaseCleanup.svelte";
	import { Fetch } from "$lib/fetchUtil";
	import { onMount } from "svelte";
	import Multifactor from "./Multifactor.svelte";
  import ResetWebPush from "./resetWebPush.svelte";

	let settings = $state({ slack: { channelID: "", enabled: false, workspace: "" }, auditlog: { enabled: false } });
	onMount(async () => {
    settings = await Fetch("/api/settings")
  })
</script>

<Slack bind:settings />
<Multifactor bind:settings />
<ResetWebPush />
<DatabaseCleanup bind:settings/>
<Export />
