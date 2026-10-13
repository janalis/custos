<?php

function optionNamesAreCaseInsensitive()
{
    setcookie("auth", "token", ["Path" => "/app", "Domain" => "example.com"]);
    setcookie("auth", "", ["expires" => 1, "path" => "/app", "domain" => "example.com"]);
}

function domainCaseIsIrrelevant()
{
    setcookie("auth", "token", ["path" => "/app", "domain" => "Example.com"]);
    setcookie("auth", "", ["expires" => 1, "path" => "/app", "domain" => "example.com"]);
}

function leadingDotIsIgnored()
{
    setcookie("auth", "token", ["path" => "/app", "domain" => ".example.com"]);
    setcookie("auth", "", ["expires" => 1, "path" => "/app", "domain" => "example.com"]);
}

function omittedPathDependsOnTheRequest()
{
    setcookie("auth", "token");
    setcookie("auth", "", 1, "/");
}

function omittedDeletionPath()
{
    setcookie("auth", "token", 0, "/app");
    setcookie("auth", "", 1);
}

function emptyPathIsNotSent()
{
    setcookie("auth", "token", ["path" => "/"]);
    setcookie("auth", "", ["expires" => 1, "path" => ""]);
}

function omittedDomain()
{
    setcookie("auth", "token", 0, "/", "example.com");
    setcookie("auth", "", 1, "/");
}

function duplicateOptionSpellings()
{
    setcookie("auth", "token", ["path" => "/app", "Path" => "/"]);
    setcookie("auth", "", ["expires" => 1, "path" => "/"]);
}

function unknownThirdArgument($options)
{
    setcookie("auth", "token", $options);
    setcookie("auth", "", 1, "/");
}

function computedOptions($extra)
{
    setcookie("auth", "token", ["path" => "/app"] + $extra);
    setcookie("auth", "", ["expires" => 1, "path" => "/"]);
}

function computedExpirySameScope()
{
    setcookie("auth", "token", time() + 3600, "/app");
    setcookie("auth", "", 1, "/app");
}

function dynamicPath($path)
{
    setcookie("auth", "token", time() + 3600, $path);
    setcookie("auth", "", 1, "/");
}
