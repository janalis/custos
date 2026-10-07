<?php
foreach ($users as $user) {
    $low[f($user)] = <warning descr="This call is repeated in the key; store its result in a local variable.">F($user)</warning>;
    $mail[$user->getEmail()] = <warning descr="This call is repeated in the key; store its result in a local variable.">$user->GetEmail()</warning>;
    $role[Acl\Role::of($user)] = <warning descr="This call is repeated in the key; store its result in a local variable.">acl\ROLE::Of($user)</warning>;
    $hash[MD5(StrToLower($user))] = md5(<warning descr="This call is repeated in the key; store its result in a local variable.">strtolower($user)</warning>);
    $prop[f($user->name)] = f($user->Name);
    $var[f($user)] = f($User);
    $const[f(LIMIT)] = f(Limit);
}
