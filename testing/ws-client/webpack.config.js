const path = require('path');

module.exports = {
	mode   : 'development',
	entry  : './src/main.ts',
	module : {
		rules : [
			{
				test    : /\.ts$/,
				use     : 'ts-loader',
				exclude : /node_modules/
			}
		]
	},
	resolve : {
		extensions : [ '.ts', '.js' ]
	},
	output : {
		filename : 'dist/bundle.js',
		path     : path.resolve(__dirname, 'dist')
	},
	devServer : {
		liveReload : false,
		port       : 3000,
		static     : {
			directory: __dirname
		}
	}
};
