<?php
function extra_length_arguments(string $s) {
    return [
        substr($s, -strlen()) === 'x',
        substr($s, -strlen('x', 2)) === 'x',
        substr($s, -strlen(...$lengths)) === 'x',
    ];
}
