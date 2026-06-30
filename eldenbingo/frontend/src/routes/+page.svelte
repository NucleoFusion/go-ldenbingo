<script lang="ts">
  import { onMount } from "svelte";
  import { IsConnected } from "$lib/wailsjs/go/main/App";
  import HomeIcon from "../components/home-icon/HomeIcon.svelte";

  let connected = $state(false);

  async function refresh() {
    connected = await IsConnected();
  }

  onMount(refresh);
</script>

<div class="grid grid-cols-[auto_1fr] gap-[1vw] mx-[2vw] mt-[2vh]">
  <div class="grid grid-rows-1 grid-cols-6 gap-4">
    <button class="grid justify-center items-center">
      <HomeIcon
        srcImg="/effigy.jpg"
        alt="connect / disconnect"
        title={connected ? "Disconnect" : "Connect"}
      />
    </button>
    <button class="grid justify-center items-center">
      <HomeIcon srcImg="/effigy.jpg" alt="Create lobby" title="Create Lobby" />
    </button>
    <button class="grid justify-center items-center">
      <HomeIcon srcImg="/finger.png" alt="join lobby" title="Join Lobby" />
    </button>
    <button disabled class="grid justify-center items-center">
      <HomeIcon srcImg="/effigy.jpg" alt="Create lobby" title="Open Map" />
    </button>
    <button disabled class="grid justify-center items-center border-green">
      <HomeIcon srcImg="/effigy.jpg" alt="Create lobby" title="Pop-Out Board" />
    </button>
    <button disabled class="grid justify-center items-center">
      <HomeIcon srcImg="/effigy.jpg" alt="Create lobby" title="Settings" />
    </button>
  </div>

  <div class="flex justify-end">
    <button class="grid justify-center items-center">
      <HomeIcon srcImg="/finger.png" alt="join lobby" title="Join Lobby" />
    </button>
  </div>
</div>
