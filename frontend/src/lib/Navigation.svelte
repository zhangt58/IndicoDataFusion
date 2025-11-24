<script>
	import { createEventDispatcher } from 'svelte';
    import { Alert } from 'flowbite-svelte';


    // export let active = 'Abstract';
	// const dispatch = createEventDispatcher();
	// function select(name) {
	// 	dispatch('navigate', { name });
	// 	active = name;
	// }
</script>


<div class="p-8">
    <Alert>
        <span class="font-medium">
            Note: This is a prototype version of the application.
            Some features may be incomplete or subject to change.
        </span>
    </Alert>
</div>

<!--<nav aria-label="Main navigation" class="navigation">-->
<!--	<ul class="flex gap-6 p-4">-->
<!--		<li>-->
<!--			<button-->
<!--				class="s-btn" class:primary={active === 'Abstract'}-->
<!--				on:click={() => select('Abstract')}-->
<!--				aria-pressed={active === 'Abstract'}-->
<!--			>-->
<!--				Abstract-->
<!--			</button>-->
<!--		</li>-->
<!--		<li>-->
<!--			<button-->
<!--				class="s-btn" class:primary={active === 'Contribution'}-->
<!--				on:click={() => select('Contribution')}-->
<!--				aria-pressed={active === 'Contribution'}-->
<!--			>-->
<!--				Contribution-->
<!--			</button>-->
<!--		</li>-->
<!--	</ul>-->
<!--</nav>-->

<style>
</style>
