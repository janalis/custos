<?php
function bounds()
{
    return [
        <error descr="Minimum is greater than maximum in this random range.">mt_rand(0b1010, 0x9)</error>,
        <error descr="Minimum is greater than maximum in this random range.">rand(0o17, 014)</error>,
        <error descr="Minimum is greater than maximum in this random range.">random_int(1_000_000, 999_999)</error>,
        <error descr="Minimum is greater than maximum in this random range.">mt_rand(3.0, 2.99)</error>,
        <error descr="Minimum is greater than maximum in this random range.">random_int(-0x10, -0x11)</error>,
        mt_rand(0x0F, 15),
        rand(1.9, 1),
        random_int(-2.5, -2),
        mt_rand(0b11, 0x10),
        random_int(1e3, 1000),
        mt_rand(1e30, 1),
    ];
}
