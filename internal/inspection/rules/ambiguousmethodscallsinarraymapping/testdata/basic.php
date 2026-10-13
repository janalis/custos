<?php
foreach ($users as $user) {
    $byEmail[$user->email()] = <warning descr="This call is repeated in the key; store its result in a local variable.">$user->email()</warning>;
    $byName[strtolower($user)] = <warning descr="This call is repeated in the key; store its result in a local variable.">strtolower($user)</warning>;
    $byName[(strtolower($user))] .= <warning descr="This call is repeated in the key; store its result in a local variable.">strtolower( $user )</warning>;
    $tagged['u:' . md5($user)] = $cache[<warning descr="This call is repeated in the key; store its result in a local variable.">md5($user)</warning>];
    $groups[Role::of($user)][] = <warning descr="This call is repeated in the key; store its result in a local variable.">Role::of($user)</warning>;
    $refs[$user?->id()] = <warning descr="This call is repeated in the key; store its result in a local variable.">$user?->id()</warning>;
}

for ($i = 0; $i < 9; $i++) {
    $sq[abs($i)] = <warning descr="This call is repeated in the key; store its result in a local variable.">abs($i)</warning>;
}

foreach ($rows as $row):
    $ids[key($row)] = <warning descr="This call is repeated in the key; store its result in a local variable.">key($row)</warning>;
endforeach;
