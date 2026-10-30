<?php
function literalVariables(string $file, string $raw, bool $cond): void
{
    // every value reaching the read is a literal that encodes
    $baseline = [['location' => 'A::b', 'count' => 1]];
    file_put_contents($file, json_encode($baseline));

    $grown = ['a' => 1];
    $grown[] = $raw;
    file_put_contents($file, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($grown)</weak_warning>);

    $either = $cond ? ['a'] : $raw;
    file_put_contents($file, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($either)</weak_warning>);

    file_put_contents($file, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($raw)</weak_warning>);
}
