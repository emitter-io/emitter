var jsonfile = require("jsonfile");
var semver = require("semver");

var buildVersion = process.env.APPVEYOR_BUILD_VERSION.substring(1);

var findPoint       = buildVersion.lastIndexOf(".");
var basePackageVer  = buildVersion.substring(0, findPoint);
var buildNumber     = buildVersion.substring(findPoint + 1, buildVersion.length);
var semversion      = semver.valid(basePackageVer + "-dev" + buildNumber);

// Set version on several projects
setVersion("../src/emitter.runtime/project.json");
setVersion("../src/emitter.server/project.json");
setVersion("../src/emitter.storage.s3/project.json");

/**
 * Set the version on the particular file. 
 * 
 * @param {any} file
 */
function setVersion(file) {
	jsonfile.readFile(file, function (err, project) {
		project.version = semversion;
		jsonfile.writeFile(file, project, {spaces: 2}, function(err) {
			console.error(err);
		});
	})
}