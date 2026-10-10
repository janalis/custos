<?php
// @noinspection ZipExtractionExceedsConfiguredBudget
$z=new ZipArchive(); $z->open("bundle.zip"); $z->extractTo("output");
