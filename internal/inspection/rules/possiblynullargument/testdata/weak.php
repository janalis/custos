<?php
// No strict_types: null to a scalar parameter of PHP's own functions is
// only deprecated; project functions still reject it.
namespace Shop;

function label(): ?string { return null; }
function shout(string $text) {}

strlen(label());
shout(<warning descr="Check this value for null before passing it to $text of Shop\shout(): it is null|string.">label()</warning>);
