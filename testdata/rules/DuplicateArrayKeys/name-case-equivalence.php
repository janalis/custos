<?php
$map = [
    'a' => Codes::Get(),
    <warning descr="Same key and value already present; remove this entry.">'a' => codes::get()</warning>,
];
$other = [
    'b' => $x->Value,
    <warning descr="Key already used earlier; the earlier entry is overwritten.">'b'</warning> => $x->value,
];
