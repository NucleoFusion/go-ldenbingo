<script lang="ts">
  let { srcImg, alt, title } = $props();
</script>

<div
  class="rounded border-[var(--accent)] bg-transparent hover:bg-[var(--accent-hover)] hover:cursor-pointer px-[1.5vw] py-[1vh]"
>
  <img class="w-[3vw]" src={srcImg} {alt} />
  <h4>
    {title}
  </h4>
</div>
