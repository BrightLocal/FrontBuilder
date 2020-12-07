[![Build Status](https://ci.bll-i.co.uk/buildStatus/icon?job=FrontBuilder/master)](https://ci.bll-i.co.uk/job/FrontBuilder/job/master)
[![Test Coverage](https://ci.bll-i.co.uk/coverage/badges/FrontBuilder/coverage_badge.svg)](https://ci.bll-i.co.uk/coverage/badges/FrontBuilder/coverage_badge.svg)

# FrontBuilder
Fast frontend assets builder/bundler

To use this library, you need to build it and place the binary in the 
root of the project or in the root of the frontend part

You also need to create a configuration file as shown in the example below:

```json
{
   "source": [
     "./templates",
     "./scripts",
     "./styles"
   ],
   "destination": "./destination",
   "index_file": "index.html",
   "html_extension": "html",
   "scripts_prefix": "js/",
   "html_prefix": "html/"
 }
```

**Required fields:**
1. `source` - define where your sources are for building;
2. `destination` - define where to put your built files;
3. `index_file` - define the path to your index.html;

**Additional settings:**
1. `html_extension` - define html extensions which used in your project;
2. `script_prefix` - if you need split your built files in different folders (look into tests-project folder). 
Where built js files would be collected in determine folder;
3. `html_prefix` - the same as script_prefix, but used for html files; 
4. `type_script_config` - define the path to tsconfig.json;

In order to inject built js files or file into html, 
it is necessary to name js and html files with the same names 
(for example look into tests-project folder).
In the html file or files you must add ```<!--#APP#-->``` 
define where to inject build script.

**Run build:**

1. `./FrontBuilder build` - run build process in `production` mode;
In this mode builder minify all js and ts files and add hash to each built js file.
Also prepare source map for each js file.
2. `./FrontBuilder build prod` - same as `build`;
3. `./FrontBuilder build dev` - run build process in `development` mode;
In this mode builder don't use any minification and all files are clear;
4. `./FrontBuilder watch` - run build process in `development` mode and start  
watching all source files for changes and rebuild project if any changes detected;
