<?php
class User {}
class Repo
{
    public function findOneBy(array $c): ?object { return null; }
    public function name(string $s): ?string { return $s === '' ? null : $s; }
}

function annotated(User $user, Repo $repo, int $n): void
{
    /** @var User $user */
    $user = $repo->findOneBy([]);
    /** @var int */
    $n = $repo->findOneBy([]);
}

function objects(User $user, Repo $repo, int $n): void
{
    $user = $repo->findOneBy([]) ?? new User();
    $n = <warning descr="Assigning a value of type \User does not match the parameter's declared type.">$repo->findOneBy([]) ?? new User()</warning>;
}

function arms(string $t, Repo $r, int $p, string $u, string $w): void
{
    $t = match ($p) {
        1 => $r->name($t),
        default => $r->name($t . 'x'),
    };
    $u = $p > 1 ? $r->name($u) : $r->name('y');
    $w = <warning descr="Assigning a value of type null does not match the parameter's declared type.">$p > 1 ? $r->name($w) : null</warning>;
    $w = <warning descr="Assigning a value of type null does not match the parameter's declared type.">match ($p) { 1 => $r->name($w), default => null }</warning>;
    $w = <warning descr="Assigning a value of type null does not match the parameter's declared type.">$r->name($w) ?: null</warning>;
}
