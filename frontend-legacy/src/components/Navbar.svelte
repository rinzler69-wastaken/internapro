<script>
  import { Link, navigate } from 'svelte-routing';
  import { auth } from '../lib/stores';

  const handleLogout = async () => {
    await auth.logout();
    navigate('/login', { replace: true });
  };

  let mobileMenuOpen = false;

  // 1. SECURITY CHECK
  // We only show the main app links if the user is logged in AND has 2FA enabled.
  $: canAccessApp = $auth.user?.is_2fa_enabled;
</script>

<nav class="bg-white border-b border-vercel-gray-200">
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
    <div class="flex justify-between h-16">
      <div class="flex">
        <div class="flex-shrink-0 flex items-center">
          <Link to="/">
            <h1 class="text-xl font-geist font-semibold">Interna</h1>
          </Link>
        </div>
        
        <div class="hidden sm:ml-8 sm:flex sm:space-x-1">
          {#if canAccessApp}
            <Link to="/dashboard" 
                  class="inline-flex items-center px-3 py-2 text-sm font-inter font-medium text-vercel-gray-700 hover:text-black transition-colors">
              Dashboard
            </Link>
            <Link to="/attendance" 
                  class="inline-flex items-center px-3 py-2 text-sm font-inter font-medium text-vercel-gray-700 hover:text-black transition-colors">
              Attendance
            </Link>
            <Link to="/tasks" 
                  class="inline-flex items-center px-3 py-2 text-sm font-inter font-medium text-vercel-gray-700 hover:text-black transition-colors">
              Tasks
            </Link>
            <Link to="/analytics" 
                  class="inline-flex items-center px-3 py-2 text-sm font-inter font-medium text-vercel-gray-700 hover:text-black transition-colors">
              Analytics
            </Link>
          {/if}

          <Link to="/profile" 
                class="inline-flex items-center px-3 py-2 text-sm font-inter font-medium text-vercel-gray-700 hover:text-black transition-colors">
            Profile
          </Link>
        </div>
      </div>

      <div class="flex items-center space-x-4">
        <div class="text-sm font-inter text-vercel-gray-600 hidden sm:block">
          {$auth.user?.email || ''}
        </div>
        <button 
          on:click={handleLogout}
          class="btn-secondary text-sm"
        >
          Logout
        </button>
      </div>

      <div class="flex items-center sm:hidden">
        <button
          on:click={() => mobileMenuOpen = !mobileMenuOpen}
          class="inline-flex items-center justify-center p-2 rounded-md text-vercel-gray-700 hover:text-black"
        >
          <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            {#if mobileMenuOpen}
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            {:else}
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
            {/if}
          </svg>
        </button>
      </div>
    </div>
  </div>

  {#if mobileMenuOpen}
    <div class="sm:hidden border-t border-vercel-gray-200">
      <div class="px-2 pt-2 pb-3 space-y-1">
        
        {#if canAccessApp}
          <Link to="/dashboard" 
                class="block px-3 py-2 text-base font-inter font-medium text-vercel-gray-700 hover:text-black hover:bg-vercel-gray-50 rounded-md transition-colors">
            Dashboard
          </Link>
          <Link to="/attendance" 
                class="block px-3 py-2 text-base font-inter font-medium text-vercel-gray-700 hover:text-black hover:bg-vercel-gray-50 rounded-md transition-colors">
            Attendance
          </Link>
          <Link to="/tasks" 
                class="block px-3 py-2 text-base font-inter font-medium text-vercel-gray-700 hover:text-black hover:bg-vercel-gray-50 rounded-md transition-colors">
            Tasks
          </Link>
          <Link to="/analytics" 
                class="block px-3 py-2 text-base font-inter font-medium text-vercel-gray-700 hover:text-black hover:bg-vercel-gray-50 rounded-md transition-colors">
            Analytics
          </Link>
        {/if}

        <Link to="/profile" 
              class="block px-3 py-2 text-base font-inter font-medium text-vercel-gray-700 hover:text-black hover:bg-vercel-gray-50 rounded-md transition-colors">
          Profile
        </Link>
      </div>
    </div>
  {/if}
</nav>