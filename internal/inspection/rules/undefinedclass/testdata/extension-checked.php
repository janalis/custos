<?php
// The constructor refuses to run without the extension, so the other
// methods use its classes knowing it is loaded.
class Signer
{
    public function __construct()
    {
        if (!extension_loaded('openacme')) {
            throw new RuntimeException('openacme extension not found');
        }
    }

    public function sign()
    {
        $handle = new OpenAcmeSign(OpenAcmeSign::ALG_SHA256);
        OpenAcme::setOption(1);
        return $handle;
    }
}

// An extension check concerns only the names it starts.
<error descr="Fix the name or install the package declaring OpenOther.">OpenOther</error>::run();

// A static method runs without an instance; a constructor that goes on
// without the extension, or checks only a stored result, guards nothing.
class StaticSigner
{
    const ALG = 1;

    public $handle = <error descr="Fix the name or install the package declaring OpenAcmeDefault.">OpenAcmeDefault</error>::ALG;

    public function __construct()
    {
        if (!extension_loaded('openacme')) {
            throw new RuntimeException('openacme extension not found');
        }
    }

    public static function make()
    {
        return new <error descr="Fix the name or install the package declaring OpenAcmeSign.">OpenAcmeSign</error>(self::ALG);
    }
}

abstract class LenientSigner
{
    abstract public function __construct();

    public function helper()
    {
    }
}

class WarningSigner
{
    public function __construct()
    {
        $loaded = extension_loaded('openacme');
        if (!extension_loaded('openacme')) {
        }
        if (!extension_loaded('openacme')) {
            trigger_error('openacme extension not found');
        }
        if (!extension_loaded('openacme')) {
            error_log('missing');
            $loaded = false;
        }
    }

    public function sign()
    {
        return new <error descr="Fix the name or install the package declaring OpenAcmeSign.">OpenAcmeSign</error>(1);
    }
}

// Returning early from a constructor still makes the instance.
class ReturningSigner
{
    public function __construct()
    {
        if (!extension_loaded('openacme')) {
            return;
        }
    }

    public function sign()
    {
        return new <error descr="Fix the name or install the package declaring OpenAcmeSign.">OpenAcmeSign</error>(1);
    }
}

class ExitingSigner
{
    public function __construct()
    {
        if (!extension_loaded('openacme')) exit(1);
    }

    public function sign()
    {
        return new OpenAcmeSign(1);
    }
}
