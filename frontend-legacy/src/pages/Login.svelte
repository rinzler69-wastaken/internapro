<script>
  
  import { auth, toast } from '../lib/stores';
  import LoadingSpinner from '../components/LoadingSpinner.svelte';

  let email = '';
  let password = '';
  let totpCode = '';
  
  let loading = false;
  let error = '';
  let showTwoFactorInput = false;

const handleSubmit = async (e) => {
    e.preventDefault();
    error = '';
    loading = true;

    try {
      const response = await auth.login(email, password, totpCode);
      
      // DEBUG: Uncomment this if you still have issues to see exactly what you're getting
      // console.log("Login Response:", response);

      // EXTRACT DATA SAFELY
      // Depending on your api.js, the actual payload might be in response.data or just response
      // We check both to be safe.
      const payload = response.data || response; 

      // 1. Check for 2FA CODE REQUEST
      if (payload.require_2fa) {
        showTwoFactorInput = true;
        loading = false;
        toast.add('Please enter your Authenticator code', 'info');
        return;
      }

      // 2. Check for SETUP REQUIREMENT (The Fix for Supervisor)
      if (payload.setup_required) {
        toast.add('Security Alert: You must set up 2FA to continue.', 'warning');
        navigate('/profile', { replace: true });
        return;
      }

      // 3. Success
      toast.add('Login successful!', 'success');
      navigate('/dashboard', { replace: true });

    } catch (err) {
      error = err.message || 'Login failed. Please check your credentials.';
      toast.add(error, 'error');

      console.error(err); // Good to see the error in console
      if (showTwoFactorInput) {
        totpCode = '';
      }
      error = err.message || 'Login failed.';
      toast.add(error, 'error');
    } finally {
      loading = false;
    }
  };
</script>

<div class="min-h-screen flex items-center justify-center bg-vercel-gray-50 py-12 px-4 sm:px-6 lg:px-8">
  <div class="max-w-md w-full space-y-8 fade-in">
    <div class="text-center">
      <h1 class="text-4xl font-geist font-bold text-black mb-2">
        Interna
      </h1>
      <p class="text-sm font-inter text-vercel-gray-600">
        Internship Management System
      </p>
    </div>

    <div class="card p-8">
      <h2 class="text-2xl font-geist font-semibold text-black mb-6">
        Sign in to your account
      </h2>

      <form on:submit={handleSubmit} class="space-y-5">
        {#if error}
          <div class="bg-red-50 border border-red-200 rounded-vercel p-3">
            <p class="text-sm font-inter text-red-700">{error}</p>
          </div>
        {/if}

        <div>
          <label for="email" class="label">Email address</label>
          <input
            id="email"
            name="email"
            type="email"
            required
            bind:value={email}
            class="input"
            placeholder="you@example.com"
            disabled={loading || showTwoFactorInput} 
          />
          </div>

        <div class={showTwoFactorInput ? "hidden" : "block"}>
          <label for="password" class="label">Password</label>
          <input
            id="password"
            name="password"
            type="password"
            required={!showTwoFactorInput}
            bind:value={password}
            class="input"
            placeholder="••••••••"
            disabled={loading}
          />
        </div>

        {#if showTwoFactorInput}
        <div class="fade-in">
            <label for="totp" class="label">Authenticator Code</label>
            <input
                id="totp"
                name="totp"
                type="text"
                required
                bind:value={totpCode}
                class="input text-center text-2xl tracking-widest font-mono"
                placeholder="000 000"
                maxlength="6"
                autocomplete="one-time-code"
                disabled={loading}
                autofocus
            />
            <p class="text-xs text-vercel-gray-500 mt-2 text-center">
                Open your Google Authenticator app and enter the code.
            </p>
        </div>
        {/if}

        <button
          type="submit"
          class="btn-primary w-full flex items-center justify-center"
          disabled={loading}
        >
          {#if loading}
            <LoadingSpinner size="sm" color="white" />
            <span class="ml-2">
                {showTwoFactorInput ? 'Verifying...' : 'Signing in...'}
            </span>
          {:else}
            {showTwoFactorInput ? 'Verify Code' : 'Sign in'}
          {/if}
        </button>
      </form>

      <div class="mt-6 text-center">
        <p class="text-sm font-inter text-vercel-gray-600">
          Don't have an account?
          <span class="text-black font-medium">
            Contact your administrator
          </span>
        </p>
      </div>
    </div>

    <div class="text-center">
      <p class="text-xs font-inter text-vercel-gray-500">
        Powered by INTERNA © 2026
      </p>
    </div>
  </div>
</div>

<style>
  :global(body) {
    background-color: #fafafa;
  }
  .hidden {
      display: none;
  }
</style>