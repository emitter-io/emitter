#region Copyright (c) 2009-2016 Misakai Ltd.
/*************************************************************************
* This program is free software: you can redistribute it and/or modify
* it under the terms of the GNU Affero General Public License as
* published by the Free Software Foundation, either version 3 of the
* License, or(at your option) any later version.
*
* This program is distributed in the hope that it will be useful,
* but WITHOUT ANY WARRANTY; without even the implied warranty of
*  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.See the
* GNU Affero General Public License for more details.
*
* You should have received a copy of the GNU Affero General Public License
* along with this program.If not, see<http://www.gnu.org/licenses/>.
*************************************************************************/
#endregion Copyright (c) 2009-2016 Misakai Ltd.

using System;
using System.IO;
using System.Net.Security;
using System.Security.Authentication;
using System.Security.Cryptography.X509Certificates;

namespace Emitter.Providers
{
    /// <summary>
    /// Represents a provider that provides transport security layer settings.
    /// </summary>
    public abstract class CertificateProvider : Provider
    {
        /// <summary>
        /// Gets or sets the certificate to use for TLS encryption/decryption.
        /// </summary>
        public abstract X509Certificate2 Certificate { get; set; }

        /// <summary>
        /// Gets or sets whether the certificate revocation check is enabled or not.
        /// </summary>
        public virtual bool CheckCertificateRevocation { get; set; }

        /// <summary>
        /// Gets or sets the password for the certificate.
        /// </summary>
        public virtual string Password { get; set; }

        /// <summary>
        /// Gets or sets the SSL/TLS protocols to use.
        /// </summary>
        public virtual SslProtocols Protocols { get; set; } = SslProtocols.Tls12 | SslProtocols.Tls11;

        /// <summary>
        /// The mode of the client certificate.
        /// </summary>
        public virtual ClientCertificateMode ClientCertificateMode { get; set; } = ClientCertificateMode.NoCertificate;

        /// <summary>
        /// Gets or sets a validation step.
        /// </summary>
        public virtual Func<X509Certificate2, X509Chain, SslPolicyErrors, bool> ClientCertificateValidation { get; set; }
    }

    /// <summary>
    /// Represents a provider that provides default implementation of transport security layer settings.
    /// </summary>
    public class FileCertificateProvider : CertificateProvider
    {
        private object Lock = new object();
        private bool X509Resolved = false;
        private X509Certificate2 X509 = null;

        /// <summary>
        /// Gets or sets the certificate to use for TLS encryption/decryption.
        /// </summary>
        public override X509Certificate2 Certificate
        {
            get
            {
                lock (this.Lock)
                {
                    if (!this.X509Resolved && this.X509 == null)
                        this.X509 = this.AutoResolveCertificate();
                    return this.X509;
                }
            }
            set
            {
                lock (this.Lock)
                {
                    this.X509 = value;
                }
            }
        }

        #region Private Members

        /// <summary>
        /// Attempts to auto-resolve and load the certificate from the data folder.
        /// </summary>
        private X509Certificate2 AutoResolveCertificate()
        {
            this.X509Resolved = true;
            try
            {
                var pfxFiles = Directory.GetFiles(Service.ConfigDirectory, "*.pfx");
                foreach (var file in pfxFiles)
                {
                    // Get the file name
                    var fileName = new FileInfo(file).Name;
                    try
                    {
                        // Load the pfx file with the password provided
                        var x509 = new X509Certificate2(file, this.Password);

                        // Construct a message
                        var message = String.Format("X509: {0} ({1}, exp. {2}) {3}",
                            String.IsNullOrWhiteSpace(x509.FriendlyName) ? fileName : x509.FriendlyName,
                            x509.SignatureAlgorithm.FriendlyName,
                            x509.NotAfter.ToString("dd/MM/yyyy"),
                            x509.HasPrivateKey ? "with private key" : String.Empty
                            );

                        Service.Logger.Log(LogLevel.Info, message);
                        return x509;
                    }
                    catch (Exception ex)
                    {
                        // Failed to load one pfx, show the error
                        Service.Logger.Log(LogLevel.Warning, String.Format("X509: Unable to load {0}. Error: {1}",
                            fileName,
                            ex.Message
                            ));
                    }
                }

                return null;
            }
            catch (Exception ex)
            {
                // Failed
                Service.Logger.Log(ex);
                return null;
            }
        }

        #endregion Private Members
    }

    public enum ClientCertificateMode
    {
        NoCertificate,
        AllowCertificate,
        RequireCertificate
    }
}