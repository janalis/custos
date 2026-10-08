<?php
final class Job
{
    /** Kept for payloads serialized before this field existed. */
    private $context = [];
    private $retry = false;
    private $queue;
    private $level;
    private $strict;
    private $tags;
    private $mode;
    private $rest;
    private $name;
    private $flag;
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

final class CorrespondenceMessage
{
    /** @var int */
    private $id;

    /** @var bool */
    private $ignoreFilled = false;

    public function __construct(int $id, bool $ignoreFilled = false)
    {
        $this->id = $id;
        $this->ignoreFilled = $ignoreFilled;
    }
}

final class SyncMessage
{
    /** @var array<string, mixed> */
    private $payload = [];

    public function __construct(int $id, bool $import = false, array $payload = [])
    {
        $this->payload = $payload;
    }
}
