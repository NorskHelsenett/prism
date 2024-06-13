<script>
	import Export from "./export.svelte";
	import Slack from "./slack.svelte";
	import DatabaseCleanup from "./databaseCleanup.svelte";
	import { Fetch } from "$lib/fetchUtil";
	import { onMount } from "svelte";
	import Multifactor from "./Multifactor.svelte";
  import ResetWebPush from "./resetWebPush.svelte";

	let settings
	onMount(async () => {
    settings = await Fetch("/api/settings")
  })
</script>

<Slack {settings} />
<Multifactor {settings} />
<ResetWebPush />
<DatabaseCleanup {settings}/>
<Export />
