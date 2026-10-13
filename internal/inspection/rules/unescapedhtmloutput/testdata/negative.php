<?php
echo '<p title="unfinished >' . $_GET['label'] . '">';
echo '<p>' . '<a title="unfinished >' . $_GET['label'] . '">';
echo '<script>' . ('<p>' . $_GET['label']);
echo '<sty' . 'le>' . ('<p>' . $_GET['label']);
echo $_GET['prefix'] . ('<p>' . $_GET['label']);
echo '<p>' . htmlspecialchars($_GET['label'], ENT_QUOTES | ENT_SUBSTITUTE, 'UTF-8') . '</p>';
