<script>
	import Export from "./export.svelte";
	import Slack from "./slack.svelte";
	import DatabaseCleanup from "./databaseCleanup.svelte";
	import { Fetch } from "$lib/fetchUtil";
	import { onMount } from "svelte";

	let settings
	onMount(async () => {
    settings = await Fetch("/api/settings")
  })
</script>

<Slack {settings} />
<DatabaseCleanup {settings}/>
<Export />
