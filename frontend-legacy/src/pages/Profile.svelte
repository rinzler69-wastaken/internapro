<script>
  import { auth, toast } from '../lib/stores';
  import { onMount } from 'svelte';
  
  // State for the modal
  let showSetupModal = false;
  let qrCodeUrl = '';
  let secretKey = '';
  let verificationCode = '';
  let loading = false;

  // 1. Start the Setup Process
  const start2FASetup = async () => {
    loading = true;
    try {
      // You'll need to ensure your api.js has a method for this, 
      // or fetch directly if you use a custom fetch wrapper
      const res = await fetch('http://localhost:8080/api/auth/2fa/setup', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}` // Assuming token storage
        }
      });
      const data = await res.json();
      
      if (res.ok) {
        qrCodeUrl = data.data.qr_code; // Backend returns { data: { qr_code: "..." } }
        secretKey = data.data.secret;
        showSetupModal = true;
      } else {
        toast.add('Failed to start 2FA setup', 'error');
      }
    } catch (e) {
      toast.add(e.message, 'error');
    } finally {
      loading = false;
    }
  };

// 2. Verify and Activate
  const confirm2FA = async () => {
    try {
      // FIX 1: Use relative URL (proxied) or your API helper
      const res = await fetch('http://localhost:8080/api/auth/2fa/verify', { 
        // Note: Ideally move this to api.js later, but for now let's fix the logic
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify({ code: verificationCode })
      });
      
      if (res.ok) {
        toast.add('2FA Enabled Successfully!', 'success');
        showSetupModal = false;

        // --- FIX 2: THE CLAUDE CATCH ---
        // We cannot use auth.update(). We must use auth.setUser().
        // We take the CURRENT user from the store ($auth.user) and modify it.
        if ($auth.user) {
            const updatedUser = { ...$auth.user, is_2fa_enabled: true };
            auth.setUser(updatedUser); // This method exists!
        }

      } else {
        toast.add('Invalid code. Please try again.', 'error');
      }
    } catch (e) {
      toast.add('Verification failed', 'error');
    }
  };

const disable2FA = async () => {
    if (!confirm('Are you sure you want to disable 2FA? This will make your account less secure.')) return;

    try {
      const res = await fetch('http://localhost:8080/api/auth/2fa/disable', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });
      
      if (res.ok) {
        toast.add('2FA has been disabled.', 'success');
        
        // --- FIX 2: THE CLAUDE CATCH ---
        if ($auth.user) {
            const updatedUser = { ...$auth.user, is_2fa_enabled: false };
            auth.setUser(updatedUser); // Using the valid method
        }

      } else {
        toast.add('Failed to disable 2FA', 'error');
      }
    } catch (e) {
      toast.add(e.message, 'error');
    }
  };
</script>

<div class="max-w-2xl mx-auto space-y-6">
  <div class="card p-6">
    <h2 class="text-xl font-geist font-semibold text-black mb-4">Security</h2>
    <p class="text-sm font-inter text-vercel-gray-600 mb-4">
      Two-factor authentication status
    </p>
    
    {#if $auth.user?.is_2fa_enabled}
      <div class="flex items-center justify-between p-4 bg-green-50 rounded-vercel border border-green-200">
      <div class="flex items-center space-x-3">
          <span class="text-sm font-inter font-medium text-green-900">2FA is enabled</span>
        </div>
        <button on:click={disable2FA} class="btn-secondary text-sm">Disable 2FA</button>
      </div>
{:else}
      <div class="flex flex-col space-y-4 p-4 bg-yellow-50 rounded-vercel border border-yellow-200">
        
        <div class="flex items-center justify-between">
            <div class="flex items-center space-x-3">
            <svg class="w-5 h-5 text-yellow-600" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
            </svg>
            <span class="text-sm font-inter font-medium text-yellow-900">2FA is not enabled</span>
            </div>
            <button on:click={start2FASetup} disabled={loading} class="btn-primary text-sm">
                {loading ? 'Loading...' : 'Enable 2FA'}
            </button>
        </div>

        <div class="text-sm text-yellow-800 bg-yellow-100 p-3 rounded border-l-4 border-yellow-500">
            <strong>Action Required:</strong> access to the Dashboard and other system features is restricted. Please enable Two-Factor Authentication to unlock full access.
        </div>

      </div>
    {/if}
  </div>
</div>

{#if showSetupModal}
<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-white p-6 rounded-lg max-w-sm w-full shadow-xl">
        <h3 class="text-lg font-bold mb-4">Setup Google Authenticator</h3>
        
        <div class="mb-4 flex justify-center">
            <img src={qrCodeUrl} alt="Scan this QR Code" class="border p-2" />
        </div>
        
        <p class="text-sm text-gray-600 mb-4 text-center">
            Scan this image with your authenticator app, then enter the 6-digit code below.
        </p>

        <input 
            type="text" 
            bind:value={verificationCode}
            placeholder="123456" 
            class="input text-center text-2xl tracking-widest mb-4" 
            maxlength="6"
        />

        <div class="flex gap-2">
            <button class="btn-secondary w-full" on:click={() => showSetupModal = false}>Cancel</button>
            <button class="btn-primary w-full" on:click={confirm2FA}>Verify & Enable</button>
        </div>
    </div>
</div>
{/if}