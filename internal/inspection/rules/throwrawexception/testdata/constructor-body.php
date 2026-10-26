<?php
class FallbackException extends \RuntimeException
{
    public function __construct($message = "", $code = 0, ?\Throwable $previous = null)
    {
        if (!$message) {
            $message = 'No sets here.';
        }
        parent::__construct($message, $code, $previous);
    }
}

class PropertyException extends \RuntimeException
{
    public function __construct($message = "", $code = 0, $previous = null)
    {
        parent::__construct($message, $code, $previous);
        $this->message = $message ?: 'Unavailable.';
    }
}

class PrefixException extends \RuntimeException
{
    public function __construct($message = "", $code = 0, $previous = null)
    {
        parent::__construct('Failed: ' . $message, $code, $previous);
    }
}

class CountedException extends \RuntimeException
{
    public function __construct($message = "", $code = 0, $previous = null)
    {
        $message++;
        parent::__construct($message, $code, $previous);
    }
}

class AliasedException extends \RuntimeException
{
    public function __construct($message = "", $code = 0, $previous = null)
    {
        $alias = &$message;
        parent::__construct($message, $code, $previous);
    }
}

class SilentException extends \RuntimeException
{
    public function __construct($message = "", $code = 0, $previous = null)
    {
    }
}

class NamedException extends \RuntimeException
{
    public function __construct($message = "", $code = 0, $previous = null)
    {
        parent::__construct(message: $message, code: $code);
    }
}

class ForwardingException extends \RuntimeException
{
    public function __construct($message = "", $code = 0, $previous = null)
    {
        $log = function () { $message = 'inner'; };
        $this->code = $code;
        parent::__construct($message, $code, $previous);
    }
}

function fail(int $n): void
{
    switch ($n) {
        case 1: throw new FallbackException();
        case 2: throw new PropertyException();
        case 3: throw new PrefixException();
        case 4: throw new CountedException();
        case 5: throw new AliasedException();
        case 6: throw new SilentException();
        case 7: throw new NamedException();
    }
    throw <weak_warning descr="Pass a message when throwing this exception.">new ForwardingException()</weak_warning>;
}
