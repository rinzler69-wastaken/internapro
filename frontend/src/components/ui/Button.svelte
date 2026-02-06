<script>
  let { 
    variant = 'primary', // primary, outline, ghost
    size = 'md',         // sm, md, lg
    children,
    class: className = '',
    ...props 
  } = $props();

  const base = "btn font-inter";
  
  const variants = {
    primary: "btn-primary",
    outline: "btn-outline",
    ghost: "btn-ghost"
  };

  const sizes = {
    sm: "btn-sm",
    md: "",
    lg: "btn-lg"
  };
</script>

<button class={`${base} ${variants[variant]} ${sizes[size]} ${className}`.trim()} {...props}>
  {@render children()}
</button>
