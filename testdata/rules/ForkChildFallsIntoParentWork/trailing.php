<?php

class ForkJob
{
    public function log(): void
    {
        echo "log";
    }

    public static function note(): void
    {
        echo "note";
    }
}

function forkNote(): void
{
    echo "note";
}

// A jump out of the child block lets the child run code the parent also runs.
function forkContinues(array $jobs)
{
    foreach ($jobs as $job) {
        $pid = pcntl_fork();
        if (<warning descr="Terminate or separate the fork child branch.">$pid === 0</warning>) {
            echo "child";
            continue;
        }
        echo "parent";
    }
}

function forkBreaks(array $jobs)
{
    foreach ($jobs as $job) {
        $pid = pcntl_fork();
        if (<warning descr="Terminate or separate the fork child branch.">$pid === 0</warning>) {
            echo "child";
            break;
        }
        echo "parent";
    }
    echo "after";
}

// A trailing call that resolves to a returning declaration falls through.
function forkReturningCalls(ForkJob $job)
{
    $pid = pcntl_fork();
    if (<warning descr="Terminate or separate the fork child branch.">$pid === 0</warning>) {
        forkNote();
    }
    $pid = pcntl_fork();
    if (<warning descr="Terminate or separate the fork child branch.">$pid === 0</warning>) {
        sleep(1);
    }
    $pid = pcntl_fork();
    if (<warning descr="Terminate or separate the fork child branch.">$pid === 0</warning>) {
        $job->log();
    }
    $pid = pcntl_fork();
    if (<warning descr="Terminate or separate the fork child branch.">$pid === 0</warning>) {
        ForkJob::note();
    }
    $pid = pcntl_fork();
    if (<warning descr="Terminate or separate the fork child branch.">$pid === 0</warning>) {
        $done = true;
    }
    $pid = pcntl_fork();
    if (<warning descr="Terminate or separate the fork child branch.">$pid === 0</warning>) {
        ;
    }
    $pid = pcntl_fork();
    if (<warning descr="Terminate or separate the fork child branch.">$pid === 0</warning>) {
        echo "child";;
    }
    return $pid;
}

function forkMaybeQuits($ok)
{
    if ($ok) {
        return;
    }
    exit(1);
}

// A declaration that can return to its caller lets the child fall through.
function forkReturningDeclaration($ok)
{
    $pid = pcntl_fork();
    if (<warning descr="Terminate or separate the fork child branch.">$pid === 0</warning>) {
        forkMaybeQuits($ok);
    }
    return $pid;
}
