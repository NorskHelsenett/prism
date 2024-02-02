<script>
  import { baseUrl } from '$lib/stores';

  export let images = []; // Prop for the list of images
  export let showModal = false; // Prop to control the visibility of the modal
  let currentImageIndex = 0; // Local state for the current image index

  // // Function to change the current image
  // function changeImage(step) {
  //   const numberOfImages = images.length;
  //   currentImageIndex = (currentImageIndex + step + numberOfImages) % numberOfImages;
  // }

  // Function to close the modal
  const closeModal = () => {
    showModal = false;
  };

    function changeImage(step) {
    currentImageIndex += step;
    if (currentImageIndex < 0) {
      currentImageIndex = images.length - 1;
    } else if (currentImageIndex >= images.length) {
      currentImageIndex = 0;
    }
  }
</script>

{#if showModal}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div class="modal modal-blur" on:click={closeModal}>
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <div class="modal-content" on:click|stopPropagation>
      <div class="">
        <button class="close-button" on:click={closeModal}>×</button>
      </div>
      <div class="">
        <!-- Carousel Implementation -->
        <div class="slide" data-bs-ride="carousel">
          <div class="carousel-indicators carousel-indicators-dot">
            {#each images as _, index}
              <button
                type="button"
                class:active={index === currentImageIndex}
                on:click={() => currentImageIndex = index}
                aria-current={index === currentImageIndex ? 'true' : 'false'}
                aria-label={`Go to slide ${index + 1}`}
              ></button>
            {/each}
          </div>
          <div class="carousel-inner">
            {#each images as image, index}
              <div class={`carousel-item ${index === currentImageIndex ? 'active' : ''}`}>
                <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
                <img src={`${$baseUrl}/api/blob/${image}`} class="d-block cursor-zoom-out" alt={`Image ${index}`} on:click={closeModal}>
              </div>
            {/each}
          </div>
          <button class="btn-carousel carousel-control-prev" type="button" on:click={() => changeImage(-1)}>
            <span class="carousel-control-prev-icon" aria-hidden="true"></span>
            <span class="visually-hidden">Previous</span>
          </button>
          <button class="btn-carousel carousel-control-next" type="button" on:click={() => changeImage(1)}>
            <span class="carousel-control-next-icon" aria-hidden="true"></span>
            <span class="visually-hidden">Next</span>
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}


<style>

.carousel-indicators-dot button {
  border: none; /* Remove default border */
  background-color: white; /* Default grey background for non-active buttons */
  width: 10px; /* Circle size */
  height: 10px; /* Circle size */
  border-radius: 50%; /* Make it round */
  margin: 0 5px; /* Spacing between circles */
  opacity: 0.5; /* Greyed out by default */
}

.carousel-indicators-dot button.active {
  background-color: #fff; /* White background for active button */
  opacity: 1; /* Fully visible */
}

.carousel-indicators {
  position: fixed;
}

.close-button {
  position: fixed;
  top: 0%;
  right: 2%;
  z-index: 999999;
  padding: 25px;
  color: #e5e4e4;
}

.carousel-item.active {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%; /* Ensure carousel item takes full height of its container */
}

.btn-carousel, .btn-carousel > span {
  position: fixed;
  color: #e5e4e4;
  filter: none;
}

.modal {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  background-color: rgba(0, 0, 0, 0.15); /* Semi-transparent background */
}

.close-button:hover{
  color: white;
}

.carousel-inner {
  width: 100%; /* Ensure carousel inner takes full width of its container */
  height: 100%; /* Adjust if you want the carousel to take full height of modal-content */
}

.carousel-inner img {
  border-radius: 5px;
  max-width: 80vw !important; /* Ensure carousel inner takes full width of its container */
  max-height: 90vh !important; /* Adjust if you want the carousel to take full height of modal-content */
}

.modal-content {
  display: flex;
  flex-direction: column;
  justify-content: center; /* Center the carousel vertically */
  align-items: center; /* Center the carousel horizontally */
  height: 75%; /* Adjust based on your preference */
  width: 75%; /* Adjust based on your preference */
  border: none;
  background-color: rgba(0,0,0,0);
  box-shadow: none;
}

.modal-header {
  padding: 10px;
  display: flex;
  justify-content: flex-end;
}

.close-button {
  border: none;
  background: none;
  cursor: pointer;
  font-size: 1.5rem;
}

.modal-body {
  display: flex;
  align-items: center;
  justify-content: space-between; /* Space out prev/next buttons and image */
  padding: 20px;
  overflow: hidden; /* Hide overflow inside the modal body */
}

.modal img {
  max-width: 100%; /* Adjust image width within modal body */
  max-height: 100%; /* Adjust image height within modal body */
  object-fit: contain; /* Ensure aspect ratio is maintained without cropping */
}

.prev-button, .next-button {
  cursor: pointer;
}

.btn-carousel{
  height: 15%;
  top: 43%;
}

</style>