<?php
function makeUser() {
    $user = new User();
    return $user->name;
}

function plain() {
    <warning descr="Variable $x is used only once; inline its value.">$x</warning> = 5;
    return $x;
}
