<?php
function render($element, $container)
{
    /** @var ElementManager $element */
    <warning descr="Variable $manager is used only once; inline its value.">$manager</warning> = $container->get('manager');
    return $manager->info($element);
}

function typed($container)
{
    /** Fetched lazily. */
    <warning descr="Variable $service is used only once; inline its value.">$service</warning> = $container->get('service');
    return $service->run();
}

function prefixed($container, $other)
{
    /** @var $other Thing */
    <warning descr="Variable $item is used only once; inline its value.">$item</warning> = $container->get('item');
    return $item->run($other);
}
