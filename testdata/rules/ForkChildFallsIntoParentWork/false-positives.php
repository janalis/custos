<?php

class ForkWorker
{
    public function stop(): never
    {
        exit(0);
    }

    public static function halt(): never
    {
        exit(0);
    }
}

function forkHalt(): never
{
    exit(0);
}

// The child block ends abruptly through a nested construct.
function forkViaFinally()
{
    $pid = pcntl_fork();
    if ($pid === 0) {
        try {
            echo "child";
        } finally {
            exit(0);
        }
    }
    pcntl_waitpid($pid, $status);
}

function forkViaIfElse($ok)
{
    $pid = pcntl_fork();
    if ($pid === 0) {
        echo "child";
        if ($ok) {
            exit(0);
        } else {
            exit(1);
        }
    }
    pcntl_waitpid($pid, $status);
}

function forkViaNestedBlock()
{
    $pid = pcntl_fork();
    if ($pid === 0) {
        echo "child";
        {
            exit(0);
        }
    }
    pcntl_waitpid($pid, $status);
}

function forkViaTrailingEmptyStatement()
{
    $pid = pcntl_fork();
    if ($pid === 0) {
        echo "child";
        exit(0);;
    }
    pcntl_waitpid($pid, $status);
}

function forkViaEndlessLoop($queue)
{
    $pid = pcntl_fork();
    if ($pid === 0) {
        while (true) {
            if ($queue->isEmpty()) {
                continue;
            }
            $queue->pop();
        }
    }
    pcntl_waitpid($pid, $status);
}

// Jumps that stay inside the child block do not carry it into parent work.
function forkWithInnerJumps(array $jobs)
{
    $pid = pcntl_fork();
    if ($pid === 0) {
        foreach ($jobs as $job) {
            if ($job === null) {
                break;
            }
            switch ($job) {
                case 1:
                    continue 2;
            }
        }
        $done = function () use ($jobs) {
            foreach ($jobs as $job) {
                break;
            }
        };
        $done();
        exit(0);
    }
    pcntl_waitpid($pid, $status);
}

// The trailing call has no resolvable declaration: it may never return.
function forkUnknownFunction($worker)
{
    $pid = pcntl_fork();
    if ($pid === 0) {
        cli_set_process_title('worker');
        runForkWorkerAndExit($worker);
    }
    return $pid;
}

function forkUnknownMethod($worker, ForkWorker $known, $name)
{
    $pid = pcntl_fork();
    if ($pid === 0) {
        $worker->runForever();
    }
    $pid = pcntl_fork();
    if ($pid === 0) {
        $known->undeclared();;
    }
    $pid = pcntl_fork();
    if ($pid === 0) {
        $known->$name();
    }
    return $pid;
}

function forkUnknownStatic($name)
{
    $pid = pcntl_fork();
    if ($pid === 0) {
        UndeclaredForkWorker::run();
    }
    $pid = pcntl_fork();
    if ($pid === 0) {
        ForkWorker::$name();
    }
    $pid = pcntl_fork();
    if ($pid === 0) {
        $name();
    }
    return $pid;
}

// The trailing call resolves to a declaration that never returns.
function forkNeverReturning(ForkWorker $worker)
{
    $pid = pcntl_fork();
    if ($pid === 0) {
        forkHalt();
    }
    $pid = pcntl_fork();
    if ($pid === 0) {
        $worker->stop();
    }
    $pid = pcntl_fork();
    if ($pid === 0) {
        ForkWorker::halt();
    }
    return $pid;
}

interface ForkRunner
{
    public function run();
}

abstract class ForkDaemon
{
    abstract public function serve();

    public function loop()
    {
        while (true) {
            sleep(1);
        }
    }

    public static function quit()
    {
        echo "bye";
        exit(0);
    }
}

function forkQuit($code)
{
    $report = function () use ($code) {
        return $code;
    };
    echo $report();
    exit($code);
}

// The implementation behind an abstract or interface method is unknown.
function forkAbstractMethods(ForkRunner $runner, ForkDaemon $daemon)
{
    $pid = pcntl_fork();
    if ($pid === 0) {
        $runner->run();
    }
    $pid = pcntl_fork();
    if ($pid === 0) {
        $daemon->serve();
    }
    return $pid;
}

// The declaration in this file ends abruptly on every path without returning.
function forkEndlessDeclarations(ForkDaemon $daemon)
{
    $pid = pcntl_fork();
    if ($pid === 0) {
        forkQuit(0);
    }
    $pid = pcntl_fork();
    if ($pid === 0) {
        $daemon->loop();
    }
    $pid = pcntl_fork();
    if ($pid === 0) {
        ForkDaemon::quit();
    }
    return $pid;
}
