<?php
class VerbPolicy
{
    /** @var uppercase-string[]|null */
    private ?array $verbs = null;

    public function fallback()
    {
        return <weak_warning descr="Operand types of '??' do not match ([array] vs [string]).">$this->verbs ?? 'all'</weak_warning>;
    }
}

function populatePolicy(array &$policy): void
{
    $policy['entries'] = ['one'];
}

function policyFallbacks(): void
{
    $policy = ['kind' => 'record'];
    populatePolicy($policy);
    // Entry types after a reference call need an output contract.
    $entries = $policy['entries'] ?? [];
    $options = $policy['options'] ?? [];
}

/** @return Broken::TYPE|null */
function uncertainPolicy() {}
function unresolvedDocFallback(): void
{
    $value = uncertainPolicy() ?? [];
}
