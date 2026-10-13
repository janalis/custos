<?php
function index(array $users, Normalizer $n) {
    $byName = [];
    foreach ($users as $user) {
        ($byName[$n->slug($user)] = <warning descr="This call is repeated in the key; store its result in a local variable.">$n->slug($user)</warning>);
        (($byName[strtolower($user)] = <warning descr="This call is repeated in the key; store its result in a local variable.">strtolower($user)</warning>));
        ($byName[$n->slug($user)] = $n->label($user));
        ($byName[next($users)] = next($users));
    }
    return $byName;
}
