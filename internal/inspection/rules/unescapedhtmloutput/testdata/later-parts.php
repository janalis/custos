<?php
// A request-derived part is found behind parts that cannot hold markup.
function laterParts($row, $flag)
{
    $name = $_GET['name'];
    $page = (int) $_GET['page'];
    $safe = htmlspecialchars($name);
    <error descr="Escape untrusted text for its HTML output context.">echo '<i>' . $safe . '</i><b>' . $name . '</b>';</error>
    <error descr="Escape untrusted text for its HTML output context.">echo "<i>$page</i><b>$name</b>";</error>
    <error descr="Escape untrusted text for its HTML output context.">echo '<i>', $page, '</i><b>', $name, '</b>';</error>
    <error descr="Escape untrusted text for its HTML output context.">print '<i>' . strlen($name) . '</i><b>' . $name</error>;
    <error descr="Escape untrusted text for its HTML output context.">echo '<a href="?page=' . $page . '">' . $name . '</a>';</error>
    <error descr="Escape untrusted text for its HTML output context.">echo '<a href="?page=', ($page + 1), '">next</a><b>', htmlentities($name), '</b><u>', $name;</error>
    <error descr="Escape untrusted text for its HTML output context.">echo <<<HTML
        <td>{$page}</td>
        <td>{$safe}</td>
        <td>{$name}</td>
        HTML;</error>
    foreach ($_GET['ids'] as $key => $id) {
        $id = (int) $id;
        <error descr="Escape untrusted text for its HTML output context.">echo "<i>$id</i><b>$key</b>";</error>
    }

    // Every part is escaped or numeric.
    echo '<i>' . $safe . '</i><b>' . $page . '</b>';
    echo "<i>$page</i><b>$safe</b>";
    // A part of unknown content may have opened an attribute or a script.
    echo '<i>' . $row['title'] . '</i><b>' . $name . '</b>';
    echo '<i>' . render($page) . '</i><b>' . $name . '</b>';
    echo "<i>{$row['title']}</i><b>$name</b>";
    // Escaped text inside a tag may end an attribute quoted the other way.
    echo "<a title='" . $safe . "'>" . $name . '</a>';
    // The request text is not in a text node.
    echo '<i>' . $page . '</i><a href="' . $name . '">';
    echo '<i>' . $safe . '</i><script>' . $name;
    // A later part of the same operand could still change the response.
    echo '<i>' . $page . '</i><b>' . $name . '</b>' . redirect();
    if ($flag) {
        $safe = load();
    }
    // Escaped on one path only: what the part holds is unknown.
    echo '<i>' . $safe . '</i><b>' . $name . '</b>';
}
