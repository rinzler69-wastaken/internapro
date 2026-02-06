<script>
  export let size = 'md'; // sm, md, lg
  export let color = 'black';

  const sizeClasses = {
    sm: 'w-4 h-4 border-2',
    md: 'w-6 h-6 border-2',
    lg: 'w-10 h-10 border-3'
  };
</script>

<div class="inline-flex items-center justify-center">
  <div class="spinner {sizeClasses[size]}" style="border-top-color: {color}"></div>
</div>
