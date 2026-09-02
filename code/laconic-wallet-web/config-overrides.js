// TODO: Use Typescript
const webpack = require('webpack')

module.exports = function override(config, env) {
  config.module.rules.push({
    test: /\.js$/,
    exclude: /node_modules[/\\](?!react-native-vector-icons)/,
    use: {
      loader: "babel-loader",
      options: {
        // Disable reading babel configuration
        babelrc: false,
        configFile: false,

        // The configuration for compilation
        presets: [
          ["@babel/preset-env", { useBuiltIns: "usage", "corejs": "3" }],
          "@babel/preset-react",
          "@babel/preset-flow",
          "@babel/preset-typescript",
        ],
        plugins: [
          "@babel/plugin-proposal-class-properties",
          "@babel/plugin-proposal-object-rest-spread",
          "@babel/plugin-transform-modules-commonjs"
        ]
      }
    }
  });

  config.plugins.push(
    new webpack.ProvidePlugin({
      Buffer: ["buffer", "Buffer"],
    }),
    require("@import-meta-env/unplugin").webpack({
      example: ".env.example",
      env: ".env",
    }),
  )

  config.module.rules.push({
    test: /\.(jpg|png|woff|woff2|eot|ttf|svg)$/,
    type: 'asset/resource'
  })

  config.resolve.fallback = {
    crypto: require.resolve("crypto-browserify"),
    stream: require.resolve("stream-browserify"),
    http: require.resolve('stream-http'),
    https: require.resolve('https-browserify'),
    url: false
  }

  config.resolve.alias['react-native$'] = require.resolve('react-native-web');

  return config;
};
