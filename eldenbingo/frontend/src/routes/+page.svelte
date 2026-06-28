<script lang="ts">
  import { onMount } from "svelte";
  import { IsConnected } from "$lib/wailsjs/go/main/App";

  let connected = false;

  async function refresh() {
    connected = await IsConnected();
  }

  onMount(refresh);
</script>

<div>
  <h1 class="text-center">Elden Bingo</h1>

  <div>
    <button> {connected ? "Disconnect" : "Connect"} </button>
    <button> Create Lobby </button>
    <button> Join Lobby </button>
    <button disabled> Open Map </button>
    <button disabled> Pop Out Board </button>
    <button> Settings </button>
  </div>
</div>
