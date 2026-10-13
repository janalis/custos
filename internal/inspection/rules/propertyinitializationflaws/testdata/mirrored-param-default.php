<?php
final class Job
{
    /** Also the value when the constructor is bypassed. */
    private $context = [];
    private $retry = false;
    private $queue = <weak_warning descr="Default is always replaced by the constructor; remove it.">'default'</weak_warning>;
    private $level = <weak_warning descr="Default is always replaced by the constructor; remove it.">1</weak_warning>;
    private $strict = <weak_warning descr="Default is always replaced by the constructor; remove it.">false</weak_warning>;
    private $tags = <weak_warning descr="Default is always replaced by the constructor; remove it.">[]</weak_warning>;
    private $mode = <weak_warning descr="Default is always replaced by the constructor; remove it.">'a'</weak_warning>;
    private $rest = <weak_warning descr="Default is always replaced by the constructor; remove it.">[]</weak_warning>;
    private $name = <weak_warning descr="Default is always replaced by the constructor; remove it.">''</weak_warning>;
    private $flag = <weak_warning descr="Default is always replaced by the constructor; remove it.">true</weak_warning>;
    private $limit = PHP_INT_MAX;

    public function __construct(
        array $context = [],
        bool $retry = false,
        string $queue = 'low',
        int $level = 1,
        bool $strict = false,
        array $tags = [],
        string $mode = 'a',
        string $name = '',
        int $limit = PHP_INT_MAX,
        ...$rest,
    ) {
        $this->context = $context;
        $this->retry = $retry;
        $this->queue = $queue;
        $this->level = $level * 2;
        $this->strict = (bool) $strict;
        $this->tags = $tags ?? [];
        $this->mode = ($mode);
        $this->rest = $rest;
        $this->name = $unknown;
        $this->flag = $retry;
        $this->limit = $limit;
    }
}

final class ReindexTask
{
    /** @var int */
    private $id;

    /** @var bool */
    private $force = false;

    public function __construct(int $id, bool $force = false)
    {
        $this->id = $id;
        $this->force = $force;
    }
}

final class PublishTask
{
    /** @var array<string, mixed> */
    private $options = [];

    public function __construct(int $id, bool $dryRun = false, array $options = [])
    {
        $this->payload = $options;
    }
}
