<script>
  import { onMount } from 'svelte';
  import { api } from '../lib/api.js';

  let notifications = $state([]);
  let loading = $state(false);

  async function fetchNotifications() {
    loading = true;
    try {
      const res = await api.getNotifications({ page: 1, limit: 20 });
      notifications = res.data || [];
    } catch (err) {
      console.error(err);
    } finally {
      loading = false;
    }
  }

  async function markRead(id) {
    await api.markNotificationRead(id);
    await fetchNotifications();
  }

  async function markAll() {
    await api.markAllNotificationsRead();
    await fetchNotifications();
  }

  onMount(fetchNotifications);
</script>

<div class="card" style="margin-bottom:16px;">
  <div style="display:flex; align-items:center; justify-content:space-between; gap:12px; flex-wrap:wrap;">
    <div>
      <h3>Notifikasi</h3>
      <p class="text-muted">Kelola notifikasi dan pembaruan terbaru.</p>
    </div>
    <button class="btn btn-outline" onclick={markAll}>Tandai Semua Dibaca</button>
  </div>
</div>

<div class="card">
  {#if loading}
    <div>Memuat...</div>
  {:else if notifications.length === 0}
    <div class="empty-state">Tidak ada notifikasi.</div>
  {:else}
    <table class="table">
      <thead>
        <tr>
          <th>Judul</th>
          <th>Pesan</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        {#each notifications as n}
          <tr>
            <td>{n.title}</td>
            <td>{n.message}</td>
            <td>
              {#if !n.read_at}
                <button class="btn btn-ghost" onclick={() => markRead(n.id)}>Tandai dibaca</button>
              {/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>
