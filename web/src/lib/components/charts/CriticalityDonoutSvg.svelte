<script>

  const cricitality = ["information", "low", "medium", "high", "critical"]
  export let severity = ""
  let index = -1

  $: {index = cricitality.indexOf(severity)}
</script>

{#if index >= 0}
  <div class="donut" style="--first: .2; --second: .2; --third: .2; --fourth: .2; --fifth: 0.2">
    {#if index >= 0}
    <div class="slice slice-first"></div>
    {/if}
    {#if index >= 1}
    <div class="slice slice-second"></div>
    {/if}
    {#if index >= 2}
    <div class="slice slice-third"></div>
    {/if}
    {#if index >= 3}
    <div class="slice slice-fourth"></div>
    {/if}
    {#if index >= 4}
    <div class="slice slice-fifth"></div>
    {/if}
    <div class="donut-label">
      <div class="donut-label-heading">
      </div>
      <div class="donut-label-sub">
      </div>
    </div>
  </div>
{/if}

<style>
  :root{
    --color-first: #4199E2;
    --color-second: #0254A5;
    --color-third: #FFA402;
    --color-fourth: #F76705;
    --color-fifth: #D7393A;
  }

  .donut {
    /* background-color: rgba(var(--tblr-primary-rgb),.13); */
  --donut-size: 40px;
  --donut-border-width: 10px;
  --donut-spacing: 0;
  --donut-spacing-color: 255, 255, 255;
  --donut-spacing-deg: calc(1deg * var(--donut-spacing));
  border-radius: 50%;
  height: var(--donut-size);
  margin: 0px;
  position: relative;
  width: var(--donut-size);
}
.donut-label {
  left: 50%;
  line-height: 1.5;
  position: absolute;
  text-align: center;
  top: 50%;
  -webkit-transform: translate(-50%, -50%);
          transform: translate(-50%, -50%);
  width: 80%;
}
.donut-label-heading {
  font-size: 24px;
  font-weight: 600;
}
.donut-label-sub {
  color: #666666;
  font-size: 14px;
  letter-spacing: 0.05em;
}
.slice {
  height: 100%;
  position: absolute;
  width: 100%;
}
.slice::before,
.slice::after {
  border: var(--donut-border-width) solid rgba(0, 0, 0, 0);
  border-radius: 50%;
  content: '';
  height: 100%;
  left: 0;
  position: absolute;
  top: 0;
  -webkit-transform: rotate(45deg);
          transform: rotate(45deg);
  width: 100%;
}
.slice::before {
  border-width: calc(var(--donut-border-width) + 1px);
  box-shadow: 0 0 1px 0 rgba(var(--donut-spacing-color), calc(100 * var(--donut-spacing)));
}
.slice-first {
  --first-start: 0;
}
.slice-first::before {
  border-top-color: rgba(var(--donut-spacing-color), calc(100 * var(--donut-spacing)));
  -webkit-transform: rotate(calc(360deg * var(--first-start) + 45deg));
          transform: rotate(calc(360deg * var(--first-start) + 45deg));
}
.slice-first::after {
  border-top-color: var(--color-first);
  border-right-color: rgba(255, 104, 56, calc(100 * (var(--first) - .25)));
  border-bottom-color: rgba(255, 104, 56, calc(100 * (var(--first) - .5)));
  border-left-color: rgba(255, 104, 56, calc(100 * (var(--first) - .75)));
  -webkit-transform: rotate(calc(360deg * var(--first-start) + 45deg + var(--donut-spacing-deg)));
          transform: rotate(calc(360deg * var(--first-start) + 45deg + var(--donut-spacing-deg)));
}
.slice-second {
  --second-start: calc(var(--first));
  --second-check: max(calc(var(--second-start) - .5), 0);
  -webkit-clip-path: inset(0 calc(50% * (var(--second-check) / var(--second-check))) 0 0);
          clip-path: inset(0 calc(50% * (var(--second-check) / var(--second-check))) 0 0);
}
.slice-second::before {
  border-top-color: rgba(var(--donut-spacing-color), calc(100 * var(--donut-spacing)));
  -webkit-transform: rotate(calc(360deg * var(--second-start) + 45deg));
          transform: rotate(calc(360deg * var(--second-start) + 45deg));
}
.slice-second::after {
  border-top-color: var(--color-second);;
  border-right-color: rgba(255, 200, 32, calc(100 * (var(--second) - .25)));
  border-bottom-color: rgba(255, 200, 32, calc(100 * (var(--second) - .5)));
  border-left-color: rgba(255, 200, 32, calc(100 * (var(--second) - .75)));
  -webkit-transform: rotate(calc(360deg * var(--second-start) + 45deg + var(--donut-spacing-deg)));
          transform: rotate(calc(360deg * var(--second-start) + 45deg + var(--donut-spacing-deg)));
}
.slice-third {
  --third-start: calc(var(--first) + var(--second));
  --third-check: max(calc(var(--third-start) - .5), 0);
  -webkit-clip-path: inset(0 calc(50% * (var(--third-check) / var(--third-check))) 0 0);
          clip-path: inset(0 calc(50% * (var(--third-check) / var(--third-check))) 0 0);
}
.slice-third::before {
  border-top-color: rgba(var(--donut-spacing-color), calc(100 * var(--donut-spacing)));
  -webkit-transform: rotate(calc(360deg * var(--third-start) + 45deg));
          transform: rotate(calc(360deg * var(--third-start) + 45deg));
}
.slice-third::after {
  border-top-color: var(--color-third);;
  border-right-color: rgba(151, 201, 92, calc(100 * (var(--third) - .25)));
  border-bottom-color: rgba(151, 201, 92, calc(100 * (var(--third) - .5)));
  border-left-color: rgba(151, 201, 92, calc(100 * (var(--third) - .75)));
  -webkit-transform: rotate(calc(360deg * var(--third-start) + 45deg + var(--donut-spacing-deg)));
          transform: rotate(calc(360deg * var(--third-start) + 45deg + var(--donut-spacing-deg)));
}
.slice-fourth {
  --fourth-start: calc(var(--first) + var(--second) + var(--third));
  --fourth-check: max(calc(var(--fourth-start) - .5), 0);
  -webkit-clip-path: inset(0 calc(50% * (var(--fourth-check) / var(--fourth-check))) 0 0);
          clip-path: inset(0 calc(50% * (var(--fourth-check) / var(--fourth-check))) 0 0);
}
.slice-fourth::before {
  border-top-color: rgba(var(--donut-spacing-color), calc(100 * var(--donut-spacing)));
  -webkit-transform: rotate(calc(360deg * var(--fourth-start) + 45deg));
          transform: rotate(calc(360deg * var(--fourth-start) + 45deg));
}
.slice-fourth::after {
  border-top-color: var(--color-fourth);;
  border-right-color: rgba(28, 178, 246, calc(100 * (var(--fourth) - .25)));
  border-bottom-color: rgba(28, 178, 246, calc(100 * (var(--fourth) - .5)));
  border-left-color: rgba(28, 178, 246, calc(100 * (var(--fourth) - .75)));
  -webkit-transform: rotate(calc(360deg * var(--fourth-start) + 45deg + var(--donut-spacing-deg)));
          transform: rotate(calc(360deg * var(--fourth-start) + 45deg + var(--donut-spacing-deg)));
}
.slice-fifth {
  --fifth-start: calc(var(--first) + var(--second) + var(--third) + var(--fourth));
  --fifth-check: max(calc(var(--fifth-start) - .5), 0);
  -webkit-clip-path: inset(0 calc(50% * (var(--fifth-check) / var(--fifth-check))) 0 0);
          clip-path: inset(0 calc(50% * (var(--fifth-check) / var(--fifth-check))) 0 0);
}
.slice-fifth::before {
  border-top-color: rgba(var(--donut-spacing-color), calc(100 * var(--donut-spacing)));
  -webkit-transform: rotate(calc(360deg * var(--fifth-start) + 45deg));
          transform: rotate(calc(360deg * var(--fifth-start) + 45deg));
}
.slice-fifth::after {
  border-top-color: var(--color-fifth);;
  border-right-color: rgba(22, 133, 184, calc(100 * (var(--fifth) - .25)));
  border-bottom-color: rgba(22, 133, 184, calc(100 * (var(--fifth) - .5)));
  border-left-color: rgba(22, 133, 184, calc(100 * (var(--fifth) - .75)));
  -webkit-transform: rotate(calc(360deg * var(--fifth-start) + 45deg + var(--donut-spacing-deg)));
          transform: rotate(calc(360deg * var(--fifth-start) + 45deg + var(--donut-spacing-deg)));
}
</style>