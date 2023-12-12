
<script>
  const title = "Project"
  const pretitle = "Project"

  function createNewProject(){}
</script>

<div class="row g-2 align-items-center">
  <div class="col">
    <div class="page-pretitle">{pretitle}</div>
    <h2 class="page-title">
      {title}
    </h2>
  </div>
  <div class="col-auto ms-auto d-print-non">
    <div class="btn-list">
      <a
        href="#"
        class="btn btn-primary d-none d-sm-inline-block"
        on:click={createNewProject}
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          class="icon"
          width="24"
          height="24"
          viewBox="0 0 24 24"
          stroke-width="2"
          stroke="currentColor"
          fill="none"
          stroke-linecap="round"
          stroke-linejoin="round"
          ><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M12 5l0 14"
          ></path><path d="M5 12l14 0"></path></svg
          >
          Add project</a
          >
        </div>
      </div>
</div>