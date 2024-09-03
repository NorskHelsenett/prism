<script>
  import { FetchFile } from '$lib/fetchUtil.js';
  import { Fetch } from '$lib/fetchUtil.js';
  import { toast } from 'svelte-sonner';

  let showModal = false;
  let modalMessage = ''; // Message to display in the modal
  let exporting = false;

  async function download() {
    exporting = true;
    try {
      await FetchFile('/api/settings/export', "prism.db");
      toast.success('Export completed successfully');
    } catch (error) {
      console.error('Export failed:', error);
      toast.error('Export failed. Please try again.');
    } finally {
      exporting = false;
    }
  }

  let fileInput;

  function uploadFile() {
    fileInput.click(); // Triggers the file selector
  }

  async function handleFileChange() {
      if (fileInput.files.length === 0) {
          console.log('No file selected');
          return;
      }

      const file = fileInput.files[0];
      // Here you can add your logic to upload the file
      upload(file)
      console.log('File selected:', file.name);
  }

  async function upload(file) {

    const formData = new FormData();
    formData.append('file', file);

    try {
      const response = await Fetch('/api/settings/import', {
        method: 'POST',
        body: formData,
      });

      if (response.message) {
        toast.success('Imported successfully');

      } else {
        toast.error('The import failed');
      }
    } catch (error) {
      toast.success(`${error}`);
    }
  }
  $: if (exporting) {
  console.log('Exporting started');
} else {
  console.log('Exporting finished');
}
</script>

<div class="card-body">
  <h2 class="card-title">Export data</h2>
  <div class="text-secondary">
    <p>Export the entire database file. The export function downloads the complete SQL file as it currently exists. This database uses Write-Ahead Logging (WAL), which offers improved concurrency and recovery. However, during the export, ensure no active write operations are occurring to prevent potential data inconsistencies. It's recommended to perform exports during periods of low database activity.</p>
    <p>Importing this file will overwrite any existing database file and terminate any active connections. Please exercise <strong>CAUTION</strong> when using this option. Before importing, ensure that all applications or services using the database are stopped to prevent data conflicts or loss. The import process replaces the current database with the uploaded file, and any unsaved changes or ongoing transactions in the old database will be lost. It's advisable to create a backup before proceeding with the import.</p>
  </div>
  <div class="btn-list justify-content-end">
    <input type="file" id="file-input" bind:this={fileInput} on:change={handleFileChange} hidden>
    <button on:click={uploadFile} class="btn btn-outline-danger" disabled={exporting}>Import Data</button>
    <button class="btn btn-primary d-none d-sm-inline-block col-2" on:click={download} disabled={exporting}>
      {#if exporting}
        <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-rotate icon-tabler icon-tabler-loader-2" width="44" height="44" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
          <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
          <path d="M12 3a9 9 0 1 0 9 9" />
        </svg>
      {/if}
      Export data</button>
  </div>
</div>

<style>
  .btn-primary[disabled] {
    border: none;
  }
</style>