<?php
function fromQuery(array $q): \DateInterval {
    $spec = $q['interval'] ?? '';
    if ($spec === '') {
        throw new \InvalidArgumentException('missing');
    }
    return new \DateInterval($spec);
}
function fromFlag(bool $long): \DateInterval {
    $spec = $long ? 'P1Y' : 'P1D';
    return new \DateInterval($spec);
}
function onlyLiteral(): \DateInterval {
    $spec = <error descr="Malformed DateInterval specification.">'P1X'</error>;
    return new \DateInterval($spec);
}
