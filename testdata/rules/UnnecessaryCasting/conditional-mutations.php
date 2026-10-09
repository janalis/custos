<?php
final class MutationConsumer
{
    public function accept($value): void {}
}

function conditionalMutationLabels(?MutationConsumer $consumer)
{
    $direct = null;
    $consumer?->accept(++$direct);
    $directLabel = '[' . (string) $direct . ']';

    $captured = null;
    $arrow = fn() => [$consumer?->accept(++$captured), '[' . (string) $captured . ']'];

    $existing = 1;
    $fallback = null;
    $existing ??= ++$fallback;
    $fallbackLabel = '[' . (string) $fallback . ']';

    $conditional = null;
    $coalescingArrow = fn() => [$existing ??= ++$conditional, '[' . (string) $conditional . ']'];
    return [$directLabel, $arrow, $fallbackLabel, $coalescingArrow];
}
