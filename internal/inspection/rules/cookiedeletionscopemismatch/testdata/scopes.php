<?php

function computedExpiry()
{
    setcookie("auth", "token", time() + 3600, "/app");
    <warning descr="Delete the cookie using its original path and domain.">setcookie("auth", "", 1, "/")</warning>;
}

function computedExpiryWithDomain($ttl)
{
    setcookie("auth", "token", $ttl, "/", "shop.example.com");
    <warning descr="Delete the cookie using its original path and domain.">setcookie("auth", "", 1, "/", "example.com")</warning>;
}

function secureFlagProvesScalarForm($ttl)
{
    setcookie(name: "auth", value: "token", expires_or_options: $ttl, secure: true, domain: "a.example.com");
    <warning descr="Delete the cookie using its original path and domain.">setcookie(name: "auth", value: "", expires_or_options: 1, domain: "b.example.com")</warning>;
}

function mixedCaseOptionNames()
{
    setcookie("auth", "token", ["Path" => "/app", "DOMAIN" => "example.com"]);
    <warning descr="Delete the cookie using its original path and domain.">setcookie("auth", "", ["Expires" => 1, "PATH" => "/", "domain" => "example.com"])</warning>;
}

function domainStillDiffersAfterFolding()
{
    setcookie("auth", "token", ["path" => "/app", "domain" => ".Shop.Example.com"]);
    <warning descr="Delete the cookie using its original path and domain.">setcookie("auth", "", ["expires" => 1, "path" => "/app", "domain" => "example.com"])</warning>;
}

function pathDiffersWhileDomainIsOmittedOnce()
{
    setrawcookie("auth", "token", 0, "/app", "example.com");
    <warning descr="Delete the cookie using its original path and domain.">setcookie("auth", "", 1, "/app/")</warning>;
}
