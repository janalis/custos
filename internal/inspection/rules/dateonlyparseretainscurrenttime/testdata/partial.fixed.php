<?php
namespace Partial;

function parse($in) {
    $r = [];
    $r[] = \DateTimeImmutable::createFromFormat('d/m', $in);
    $r[] = \DateTimeImmutable::createFromFormat('F Y', $in);
    $r[] = \DateTime::createFromFormat('Y', $in);
    $r[] = \DateTime::createFromFormat('m', $in);
    $r[] = \DateTime::createFromFormat('Y-m', $in);
    $r[] = \DateTime::createFromFormat('z Y', $in);
    $r[] = \DateTime::createFromFormat('Y-m-\\d', $in);
    $r[] = \DateTimeImmutable::createFromFormat('!y n j', $in);
    $r[] = \DateTimeImmutable::createFromFormat("!j F Y", $in);
    $r[] = \DateTimeImmutable::createFromFormat('!D, d M Y', $in);
    $r[] = \DateTimeImmutable::createFromFormat('!Y-m-d', $in)->format('Y-m-d');
    $r[] = \DateTimeImmutable::createFromFormat('!Y-m-d', $in)->setDate(2026, 1, 1);
    $d = \DateTimeImmutable::createFromFormat('!Y-m-d', $in);
    $r[] = $d->setTime(0, 0);
    $r[] = $in->setTime(\DateTime::createFromFormat('!Y-m-d', $in), 0);
    $m = 'setTime';
    $r[] = \DateTime::createFromFormat('!Y-m-d', $in)->$m(0, 0);
    return $r;
}
