<script lang="ts">
  import { onMount } from "svelte";
  // import { IsConnected } from "$lib/wailsjs/go/main/App";
  import HomeIcon from "../components/home-icon/HomeIcon.svelte";
  import Console from "../components/console/console.svelte";
  import { consoleStore } from "$lib/stores/console.svelte";

  let connected = $state(false);

  // async function refresh() {
  //   connected = await IsConnected();
  // }

  // onMount(refresh);
  onMount(() => {
    consoleStore.add("info", "Connecting to 23.95.170.222:4501...");

    setInterval(() => {
      consoleStore.add("info", "Connected to server");
    }, 1000);

    setInterval(() => {
      consoleStore.add("warn", "Connection is unstable");
    }, 2000);
  });
</script>

<div class="grid grid-rows-[auto_1fr_auto] w-full h-[100vh]">
  <div class="grid grid-cols-[auto_1fr] gap-[1vw] p-1">
    <div class="flex gap-4">
      <button class="grid justify-center items-center">
        <HomeIcon
          srcImg={connected
            ? "/_connectButton.Image.png"
            : "/_disconnectButton.Image.png"}
          alt="connect / disconnect"
          title={connected ? "Disconnect" : "Connect"}
        />
      </button>
      <button class="grid justify-center items-center">
        <HomeIcon
          srcImg="/_createLobbyButton.Image.png"
          alt="Create lobby"
          title="Create Lobby"
        />
      </button>
      <button class="grid justify-center items-center">
        <HomeIcon
          srcImg="/_joinLobbyButton.Image.png"
          alt="join lobby"
          title="Join Lobby"
        />
      </button>
      <button disabled class="grid justify-center items-center">
        <HomeIcon
          srcImg="/_openMapButton.Image.png"
          alt="Open Map"
          title="Open Map"
        />
      </button>
      <button disabled class="grid justify-center items-center border-green">
        <HomeIcon
          srcImg="/_openExternalBoardToolStripButton.Image.png"
          alt="Pop-Out Board"
          title="Pop-Out Board"
        />
      </button>
      <button disabled class="grid justify-center items-center">
        <HomeIcon
          srcImg="/_settingsButton.Image.png"
          alt="Settings"
          title="Settings"
        />
      </button>
    </div>

    <div class="flex justify-end">
      <button class="grid justify-center items-center">
        <HomeIcon
          srcImg="/_startGameButton.Image.png"
          alt="join lobby"
          title="Start Elden Ring"
        />
      </button>
    </div>
  </div>

  <!-- Console -->
  <Console />

  <!-- Status Bar -->
  <!-- TODO: Add logic -->
  <div class="flex p-2">
    <p class="flex-1 text-center">Connected - Not in Lobby</p>
    <span class="px-4 bg-green">Waiting for Game</span>
  </div>
</div>
