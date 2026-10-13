<?php
function relay(array $input, bool $multi)
{
    $reader = 'parse_str';
    if ($multi) {
        $reader .= '_multi';
    }
    $reader($input, $out);
    return array_map($reader, $input);
}
