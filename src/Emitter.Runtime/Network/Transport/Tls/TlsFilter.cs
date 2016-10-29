using System;
using System.Net.Security;
using System.Security.Cryptography.X509Certificates;
using System.Threading.Tasks;
using Emitter.Providers;

namespace Emitter.Network.Tls
{
    /// <summary>
    /// Represents a TLS/SSL connection filter.
    /// </summary>
    public class TlsFilter : IConnectionFilter
    {
        private readonly CertificateProvider Options;
        private readonly IConnectionFilter Previous;

        /// <summary>
        /// Creates a new instance of an object.
        /// </summary>
        /// <param name="previous">Previous connection filter to apply.</param>
        public TlsFilter(IConnectionFilter previous = null)
        {
            if (previous == null)
                previous = new NoOpConnectionFilter();

            this.Options = Service.Providers.Resolve<CertificateProvider>();
            this.Previous = previous;
        }

        /// <summary>
        /// Occurs when a connection is established.
        /// </summary>
        /// <param name="context"></param>
        /// <returns></returns>
        public async Task OnConnectionAsync(ConnectionFilterContext context)
        {
            await this.Previous.OnConnectionAsync(context);

            if (string.Equals(context.Address.Scheme, "https", StringComparison.OrdinalIgnoreCase))
            {
                SslStream sslStream;
                if (this.Options.ClientCertificateMode == ClientCertificateMode.NoCertificate)
                {
                    sslStream = new SslStream(context.Connection);
                    await sslStream.AuthenticateAsServerAsync(this.Options.Certificate, clientCertificateRequired: false,
                        enabledSslProtocols: this.Options.Protocols, checkCertificateRevocation: this.Options.CheckCertificateRevocation);
                }
                else
                {
                    sslStream = new SslStream(context.Connection, leaveInnerStreamOpen: false,
                        userCertificateValidationCallback: (sender, certificate, chain, sslPolicyErrors) =>
                        {
                            if (certificate == null)
                                return this.Options.ClientCertificateMode != ClientCertificateMode.RequireCertificate;

                            if (this.Options.ClientCertificateValidation == null)
                            {
                                if (sslPolicyErrors != SslPolicyErrors.None)
                                    return false;
                            }

                            var certificate2 = ConvertToX509Certificate2(certificate);
                            if (certificate2 == null)
                                return false;

                            if (this.Options.ClientCertificateValidation != null)
                            {
                                if (!this.Options.ClientCertificateValidation(certificate2, chain, sslPolicyErrors))
                                    return false;
                            }

                            return true;
                        });

                    await sslStream.AuthenticateAsServerAsync(
                        this.Options.Certificate,
                        clientCertificateRequired: true,
                        enabledSslProtocols: this.Options.Protocols,
                        checkCertificateRevocation: this.Options.CheckCertificateRevocation
                        );
                }

                // Promote the connection to an SSL Stream
                context.Connection = sslStream;
            }
        }

        /// <summary>
        /// Converts a certificate to a X509Certificate2.
        /// </summary>
        /// <param name="certificate">The certificate to convert.</param>
        /// <returns></returns>
        private X509Certificate2 ConvertToX509Certificate2(X509Certificate certificate)
        {
            if (certificate == null)
                return null;

            X509Certificate2 certificate2 = certificate as X509Certificate2;
            if (certificate2 != null)
                return certificate2;

#if DOTNET
            // conversion X509Certificate to X509Certificate2 not supported
            // https://github.com/dotnet/corefx/issues/4510
            return null;
#else
            return new X509Certificate2(certificate);
#endif
        }
    }
}