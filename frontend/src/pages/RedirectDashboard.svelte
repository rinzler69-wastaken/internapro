<script>
  import { onMount } from 'svelte';
  import { replace } from '@mateothegreat/svelte5-router';

  onMount(() => {
    replace('/dashboard');
  });
</script>

<div class="main-content">
  <div class="card">Mengalihkan ke dashboard...</div>
</div>
