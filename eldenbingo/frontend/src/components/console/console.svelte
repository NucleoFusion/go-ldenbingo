<script lang="ts">
  import { consoleStore } from "$lib/stores/console.svelte";
  import { tick } from "svelte";
  let container: HTMLDivElement;

  $effect(() => {
    consoleStore.logs.length;

    (async () => {
      await tick();

      if (container) {
        container.scrollTop = container.scrollHeight;
      }
    })();
  });
</script>

<div
  bind:this={container}
  class="bg-[#1e1e2e] mx-[1vw] p-1 overflow-auto min-h-0 h-full"
>
  {#each consoleStore.logs as log}
    <p class="text-[var(--text)]" class:text-red-500={log.level == "warn"}>
      {`[${log.time}] ${log.message}`}
    </p>
  {/each}
</div>
