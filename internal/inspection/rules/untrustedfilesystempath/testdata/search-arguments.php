<?php
function export(string $path, string $dir)
{
    // What is searched for is removed from the result, never copied in.
    $file = str_replace($_SERVER['REQUEST_URI'] . $dir, 'public:/', $path);
    file_get_contents($file);
    file_get_contents(str_ireplace(search: $_GET['q'], replace: 'x', subject: $path));
    file_get_contents(preg_replace('/' . $_GET['q'] . '/', '', $path));
    // The replacement and the subject are copied.
    <error descr="Validate filesystem containment before using this path.">file_get_contents(str_replace('x', $_GET['q'], $path))</error>;
    <error descr="Validate filesystem containment before using this path.">file_get_contents(str_replace($dir, 'x', $_GET['q']))</error>;
    <error descr="Validate filesystem containment before using this path.">file_get_contents(str_replace(subject: $path, search: 'a', replace: $_GET['q']))</error>;
    // Unpacked arguments are not followed.
    file_get_contents(str_replace(...$_GET['parts']));
}
