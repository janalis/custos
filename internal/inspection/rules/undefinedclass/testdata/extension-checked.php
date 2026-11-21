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

// Only a missing extension the constructor refuses counts: not a version
// check, nor a check refusing the extension when it is loaded, and not in
// the constructor itself or a function declared in a method.
class VersionSigner
{
    public function __construct()
    {
        if (PHP_VERSION_ID < 70000) {
            throw new RuntimeException('PHP 7 required');
        }
        if (extension_loaded('openacme')) {
            throw new RuntimeException('openacme conflicts');
        }
    }

    public function sign()
    {
        return new <error descr="Fix the name or install the package declaring OpenAcmeSign.">OpenAcmeSign</error>(1);
    }
}

class EitherSigner
{
    public function __construct()
    {
        $early = new <error descr="Fix the name or install the package declaring OpenAcmeSign.">OpenAcmeSign</error>(0);
        if (PHP_VERSION_ID < 50600 || !extension_loaded('openacme')) {
            throw new RuntimeException('openacme extension not found');
        }
    }

    public function sign()
    {
        function openacme_helper()
        {
            return new <error descr="Fix the name or install the package declaring OpenAcmeSign.">OpenAcmeSign</error>(2);
        }
        return new OpenAcmeSign(1);
    }
}

// An extension starts a name only up to a word boundary.
if (extension_loaded('opena')) {
    new <error descr="Fix the name or install the package declaring OpenAcme.">OpenAcme</error>();
}
if (extension_loaded('')) {
    new <error descr="Fix the name or install the package declaring OpenEmpty.">OpenEmpty</error>();
}

class ParenSigner
{
    public function __construct()
    {
        if ((!extension_loaded('openacme'))) {
            throw new RuntimeException('openacme extension not found');
        }
    }

    public function sign()
    {
        return new OpenAcmeSign(1);
    }
}
