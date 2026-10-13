<?php
namespace Partial;

function parse($in) {
    $r = [];
    $r[] = <warning descr="Reset omitted time fields when parsing a date-only value.">\DateTimeImmutable::createFromFormat('d/m', $in)</warning>;
    $r[] = <warning descr="Reset omitted time fields when parsing a date-only value.">\DateTimeImmutable::createFromFormat('F Y', $in)</warning>;
    $r[] = <warning descr="Reset omitted time fields when parsing a date-only value.">\DateTime::createFromFormat('Y', $in)</warning>;
    $r[] = <warning descr="Reset omitted time fields when parsing a date-only value.">\DateTime::createFromFormat('m', $in)</warning>;
    $r[] = <warning descr="Reset omitted time fields when parsing a date-only value.">\DateTime::createFromFormat('Y-m', $in)</warning>;
    $r[] = <warning descr="Reset omitted time fields when parsing a date-only value.">\DateTime::createFromFormat('z Y', $in)</warning>;
    $r[] = <warning descr="Reset omitted time fields when parsing a date-only value.">\DateTime::createFromFormat('Y-m-\\d', $in)</warning>;
    $r[] = <warning descr="Reset omitted time fields when parsing a date-only value.">\DateTimeImmutable::createFromFormat('y n j', $in)</warning>;
    $r[] = <warning descr="Reset omitted time fields when parsing a date-only value.">\DateTimeImmutable::createFromFormat("j F Y", $in)</warning>;
    $r[] = <warning descr="Reset omitted time fields when parsing a date-only value.">\DateTimeImmutable::createFromFormat('D, d M Y', $in)</warning>;
    $r[] = <warning descr="Reset omitted time fields when parsing a date-only value.">\DateTimeImmutable::createFromFormat('Y-m-d', $in)</warning>->format('Y-m-d');
    $r[] = <warning descr="Reset omitted time fields when parsing a date-only value.">\DateTimeImmutable::createFromFormat('Y-m-d', $in)</warning>->setDate(2026, 1, 1);
    $d = <warning descr="Reset omitted time fields when parsing a date-only value.">\DateTimeImmutable::createFromFormat('Y-m-d', $in)</warning>;
    $r[] = $d->setTime(0, 0);
    $r[] = $in->setTime(<warning descr="Reset omitted time fields when parsing a date-only value.">\DateTime::createFromFormat('Y-m-d', $in)</warning>, 0);
    $m = 'setTime';
    $r[] = <warning descr="Reset omitted time fields when parsing a date-only value.">\DateTime::createFromFormat('Y-m-d', $in)</warning>->$m(0, 0);
    return $r;
}
