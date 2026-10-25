<?php
function render($element, $container)
{
    /** @var ElementManager $element */
    return $container->get('manager')->info($element);
}

function typed($container)
{
    return $container->get('service')->run();
}

function prefixed($container, $other)
{
    /** @var $other Thing */
    return $container->get('item')->run($other);
}
