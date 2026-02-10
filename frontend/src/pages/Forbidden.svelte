<script>
  import { route } from '@mateothegreat/svelte5-router';
</script>

<div class="card">
  <h3>403 - Akses Ditolak</h3>
  <p class="text-muted">
    Anda tidak memiliki izin untuk membuka halaman ini.
  </p>
  <a href="/dashboard" use:route class="btn btn-outline" style="margin-top:12px;">
    Kembali ke Dashboard
  </a>
</div>
