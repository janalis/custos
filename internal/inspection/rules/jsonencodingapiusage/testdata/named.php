<?php
namespace App\Settings;

// The file writes global constants qualified (\PHP_EOL): so does the fix.
function rows(string $raw, array $v, bool $pretty, int $a, int $b): array
{
    return [
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">\json_decode($raw, associative: true)</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($v, JSON_UNESCAPED_SLASHES)</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">\json_encode($v, flags: $pretty ? \JSON_PRETTY_PRINT : 0)</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">\json_encode($v, $pretty ? \JSON_PRETTY_PRINT : 0)</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">\json_encode($v, $a | $b)</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">\json_encode(value: $v)</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">\json_encode($v, $a ?: $b)</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">\json_encode($v, $a ?? $b)</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">\json_decode($raw, depth: 8, associative: true)</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">\json_decode(
            $raw,
            associative: false,
        )</weak_warning>,
        \PHP_EOL,
    ];
}
