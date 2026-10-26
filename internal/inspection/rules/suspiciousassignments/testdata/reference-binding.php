<?php
function build(array &$byKey, ?string $key, array $data): array
{
    if (null !== $key) {
        $output = &$byKey[$key];
    }
    $output = $data;

    if ($key === 'x') {
        $copy = $data;
    }
    <error descr="$copy is overwritten right after the 'if'; an 'else' may be missing.">$copy = []</error>;

    return [$output, $copy];
}
