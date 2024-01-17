<script>
    import { FetchFile } from '$lib/fetchUtil.js';
    import { Fetch } from '$lib/fetchUtil.js';
  import { notification } from '$lib/stores/notificationStore';

    let showModal = false;
    let modalMessage = ''; // Message to display in the modal

    function download() {
      FetchFile('/api/settings/export', "prism.db");
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
          notification.addAlert({
            type: 'success',
            title: 'Imported successfully',
            message: response.message
          });
        } else {
          notification.addAlert({
            type: 'warning',
            title: 'The import failed',
            message: response.error
          });
        }
      } catch (error) {
        console.error('Error uploading file:', error);
          notification.addAlert({
            type: 'danger',
            title: 'Importing',
            message: error
          });
      }
    }
</script>
<div class="card-body">
  <h2 class="card-title">Export data</h2>
  <p class="text-secondary">Export the entire database file. Importing this file will overwrite any existing database file and terminate any active connections. All users will be <strong>logged out.</strong> Please exercise <strong >CAUTION</strong> when using this option. </p>
  <span class="badge bg-orange-lt mb-3">The server will perform a hard restart after the import</span>
  <div class="btn-list justify-content-end">
    <input type="file" id="file-input" bind:this={fileInput} on:change={handleFileChange} hidden>
    <button on:click={uploadFile} class="btn btn-outline-danger">Import Data</button>
    <button class="btn btn-primary d-none d-sm-inline-block" on:click={download}>Export data</button>
  </div>
</div>
