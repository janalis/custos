<?php
// Properties listed by __sleep() are serialised.
class Connection
{
    private $dsn;
    private $user;
    private $handle;

    public function __construct($dsn, $user)
    {
        $this->dsn = $dsn;
        $this->user = $user;
        <weak_warning descr="Private property is only used in the constructor; likely dead code.">$this->handle</weak_warning> = null;
    }

    public function __sleep()
    {
        $names = ['dsn', 'other', 3];
        $names[] = "\0Connection\0user";
        return $names;
    }
}

trait Sleeps
{
    public function __sleep() { return ['token']; }
}

class Session
{
    use Sleeps;

    private $token;

    public function __construct($token) { $this->token = $token; }
}

// A name in any other method is only a string.
class Labelled
{
    private $label;

    public function __construct($label) { <weak_warning descr="Private property is only used in the constructor; likely dead code.">$this->label</weak_warning> = $label; }

    public function names() { return ['label']; }
}
