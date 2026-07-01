<script lang="ts">
  let { srcImg, alt, title } = $props();
</script>

<div
  class="rounded border-[var(--accent)] bg-transparent hover:bg-black/20 hover:cursor-pointer"
>
  <div class="grid items-center justify-center">
    <img class="h-[50px]" src={srcImg} {alt} />
  </div>
  <h4 class="text-sm pt-1">
    {title}
  </h4>
</div>
